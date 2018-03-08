package scheduler

import (
	"sys/fetch/base"
	"sync"
	"fmt"
)

//请求缓存的接口类型
type requestCache interface {
	//将请求放入请求缓存
	put(req *base.Request) bool
	//从请求缓存获取最早被放入且人在其中的请求
	get() *base.Request
	//获得请求缓存的容量
	capacity() int
	//获得请求缓存的实时长度，即其中的请求的即时数量
	length() int
	//关闭请求缓存
	close()
	//获取请求缓存的摘要信息
	summary() string
}

type reqCacheBySlice struct {
	cache  []*base.Request //缓存
	mutex  sync.Mutex      //锁
	status byte            //状态 0:表示请求缓存正在运行 1:表示请求缓存已被关闭
}

var statusMap = map[byte]string {
	0: "running",
	1: "closed",
}

//创建请求缓存
func newRequestCache() requestCache {
	rc := &reqCacheBySlice{
		cache: make([]*base.Request, 0),
	}
	return rc
}

func (rcache *reqCacheBySlice) put(req *base.Request) bool {
	if req == nil {
		return false
	}
	if rcache.status == 1 {
		return false
	}
	rcache.mutex.Lock()
	defer rcache.mutex.Unlock()
	rcache.cache = append(rcache.cache, req)
	return true
}

func (rcache *reqCacheBySlice) get() *base.Request {
	if rcache.length() == 0 {
		return nil
	}
	if rcache.status == 1 {
		return nil
	}
	rcache.mutex.Lock()
	defer rcache.mutex.Unlock()
	req := rcache.cache[0]
	rcache.cache = rcache.cache[1:]
	return req
}

func (rcache *reqCacheBySlice) capacity() int {
	return cap(rcache.cache)
}

func (rcache *reqCacheBySlice) length() int {
	return len(rcache.cache)
}

func (rcache *reqCacheBySlice) close() {
	if rcache.status == 1 {
		return
	}
	rcache.status = 1
}

//摘要信息模板
var summaryTemplate = "status: %s," + "length: %d," + "capacity: %d"

func (rcache *reqCacheBySlice) summary() string {
	summary := fmt.Sprintf(summaryTemplate,
		statusMap[rcache.status],
		rcache.length(),
		rcache.capacity())
	return summary
}
