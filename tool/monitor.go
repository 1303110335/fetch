package tool

import (
	sched "sys/fetch/scheduler"
	"time"
	"errors"
	"fmt"
	"runtime"
)

//日志记录函数的类型
//参数level代表日志级别。级别设定：0：普通；1：警告；2：错误
type Record func(level byte, content string)

//调度监控函数
func Monitoring(
	scheduler sched.Scheduler, //代表作为监控目标的调度器
	intervalNs time.Duration,  //检查时间间隔
	maxIdleCount uint,         //最大空闲计数
	autoStop bool,             //表示该方法是否在调度器空闲一段时间之后自行停止调度器
	detailSummary bool,        //是否需要详细的摘要信息
	record Record) <-chan uint64 {

	if scheduler == nil {
		panic(errors.New("The scheduler is invalid!"))
	}
	//繁殖过小的参数值对爬取流程的影响
	if intervalNs < time.Millisecond {
		intervalNs = time.Millisecond
	}
	if maxIdleCount < 1000 {
		maxIdleCount = 1000
	}
	//监控停止通知器
	stopNotifier := make(chan byte, 1)
	//接收和报告错误
	reportError(scheduler, record, stopNotifier)
	//记录摘要信息
	recordSummary(scheduler, detailSummary, record, stopNotifier)
	//检查计数通道
	checkCountChan := make(chan uint64, 2)
	//检查空闲状态
	checkStatus(scheduler,
		intervalNs,
		maxIdleCount,
		autoStop,
		checkCountChan,
		record,
		stopNotifier)
	return checkCountChan
}

//接收和报告错误
func reportError(
	scheduler sched.Scheduler,
	record Record,
	stopNotifier <-chan byte) {
	go func() {
		//等待调度器开启
		waitForSchedulerStart(scheduler)
		for {
			//查看监控停止通知
			select {
			case <-stopNotifier:
				return
			default:
			}
			errorChan := scheduler.ErrorChan()
			if errorChan == nil {
				return
			}
			err := <-errorChan
			if err != nil {
				errMsg := fmt.Sprintf("Error (received from error channel): %s", err)
				record(2, errMsg)
			}
			time.Sleep(time.Millisecond)
		}
	}()
}

//等待调度器开启
func waitForSchedulerStart(scheduler sched.Scheduler) {
	for !scheduler.Running() {
		time.Sleep(time.Millisecond)
	}
}

//摘要信息的模板
var summaryForMonitoring = "Monitor - COllected information[%d]:\n" +
	"	Goroutine number : %d\n" +
	"	Scheduler:\n %s" +
	"Escaped time: %s\n"

//已达到直达空闲计数的消息模板
var msgReachMaxIdleCount = "The scheduler has been idle for a period of time" +
	"(about %s)." +
	" Now consider what stop it."

// 停止调度器的消息模板。
var msgStopScheduler = "Stop scheduler...%s."

func recordSummary(
	scheduler sched.Scheduler,
	detailSummary bool,
	record Record,
	stopNotifier <-chan byte) {
	var recordCount uint64 = 1
	startTime := time.Now()
	var prevSchedSummary sched.SchedSummary
	var prevNumGoroutine int

	for {
		//产看监控停止通知器
		select {
		case <-stopNotifier:
			return
		default:
		}
		//获取摘要信息的各组成部分
		currNumGoroutine := runtime.NumGoroutine()
		currSchedSummary := scheduler.Summary("   ")
		if currNumGoroutine != prevNumGoroutine || !currSchedSummary.Same(prevSchedSummary) {
			schedSummaryStr := func() string {
				if detailSummary {
					return currSchedSummary.Detail()
				} else {
					return currSchedSummary.String()
				}
			}()
			//记录摘要信息
			info := fmt.Sprintf(summaryForMonitoring,
				recordCount,
				currNumGoroutine,
				schedSummaryStr,
				time.Since(startTime).String(),
			)
			record(0, info)
			prevNumGoroutine = currNumGoroutine
			prevSchedSummary = currSchedSummary
			recordCount++
		}
		time.Sleep(time.Millisecond)
	}
}

func checkStatus(
	scheduler sched.Scheduler,
	intervalNs time.Duration,
	maxIdleCount uint,
	autoStop bool,
	checkCountChan chan<- uint64,
	record Record,
	stopNotifier chan<- byte) {
	var checkCount uint64
	go func() {
		defer func() {
			stopNotifier <- 1
			stopNotifier <- 2
			checkCountChan <- checkCount
		}()
		var idleCount uint //连续空闲状态的计数
		var firstIdleTime time.Time

		for {
			//检查调度器的空闲状态
			if scheduler.Idle() {
				idleCount ++
				if idleCount == 1 {
					firstIdleTime = time.Now()
				}
				if idleCount >= maxIdleCount {
					msg := fmt.Sprintf(msgReachMaxIdleCount, time.Since(firstIdleTime).String())
					record(0, msg)
					//再次检查调度器的空闲状态，确保它已经可以被停止
					if scheduler.Idle() {
						if autoStop {
							var result string
							if scheduler.Stop() {
								result = "success"
							} else {
								result = "failing"
							}
							msg = fmt.Sprintf(msgStopScheduler, result)
							record(0, msg)
						}
						break
					} else {
						if idleCount > 0 {
							idleCount = 0
						}
					}
				}
			} else {
				if idleCount > 0 {
					idleCount = 0
				}
			}
			checkCount++
			time.Sleep(intervalNs)
		}
	}()
}
