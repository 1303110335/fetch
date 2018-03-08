package middleware

import (
	"sync"
	"math"
)

type IdGenerator interface {
	GetUint32() uint32	//获得一个uint32类型的ID
}

// 创建ID生成器。
func NewIdGenerator() IdGenerator {
	return &cyclicIdGenerator{}
}

type cyclicIdGenerator struct {
	sn 			uint32 	//当前的ID
	ended 		bool 	//前一个ID是否已经为其类型所能表示的最大值
	mutex 		sync.Mutex 	//互斥锁
}

func (gen *cyclicIdGenerator) GetUint32() uint32 {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()
	if gen.ended {
		defer func() { gen.ended = false }()
		gen.sn = 0
		return gen.sn
	}
	id := gen.sn
	if id < math.MaxUint32 {
		gen.sn ++
	} else {
		gen.ended = true
	}
	return id
}