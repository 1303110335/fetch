package middleware

import (
	"sync"
	"fmt"
)

//停止信号的接口类型
type StopSign interface {
	//置位停止信号
	Sign() bool
	//判断停止信号是否已被发出
	Signed() bool
	//充值停止信号
	Reset()
	//处理停止信号
	Deal(code string)
	//获取某一个停止信号处理方的处理计数
	DealCount(code string) uint32
	//获取停止信号的被停止的总计数
	DealTotal() uint32
	//获取摘要信息
	Summary() string
}

func NewStopSign() StopSign {
	ss := &myStopSign{
		dealCountMap: make(map[string]uint32),
	}
	return ss
}

type myStopSign struct {
	signed bool 					//表示信号是否已发出
	dealCountMap map[string]uint32 	//处理计数的字典
	rwmutex sync.RWMutex			//读写锁
}

func (ss *myStopSign) Sign() bool {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	if ss.signed {
		return false
	}
	ss.signed = true
	return true
}

func (ss *myStopSign) Signed() bool {
	return ss.signed
}

func (ss *myStopSign) Deal(code string) {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	if !ss.signed{
		return
	}
	if _, ok := ss.dealCountMap[code]; !ok {
		ss.dealCountMap[code] = 1
	} else {
		ss.dealCountMap[code] += 1
	}
}

func (ss *myStopSign) Reset() {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	ss.signed = false
	ss.dealCountMap = make(map[string]uint32)
}

func (ss *myStopSign) DealCount(code string) uint32 {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	return ss.dealCountMap[code]
}

func (ss *myStopSign) DealTotal() uint32 {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	var total uint32
	for _, v := range ss.dealCountMap {
		total += v
	}
	return total
}

func (ss *myStopSign) Summary() string {
	if ss.signed {
		return fmt.Sprintf("signed: true, dealCount: %v", ss.dealCountMap)
	} else {
		return "signed: false"
	}
}