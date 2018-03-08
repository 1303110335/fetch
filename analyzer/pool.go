package analyzer

import (
	"reflect"
	mdw "sys/fetch/middleware"
	"fmt"
	"errors"
)

//生成分析器的函数类型
type GenAnalyzer func() Analyzer

//分析器池的接口类型
type AnalyzerPool interface {
	Take() (Analyzer, error)  //从池中取出一个分析器
	Return(analyzer Analyzer) error //把一个分析器归还给池
	Total() uint32					//获得池的总量
	Used() uint32					//获得正在被使用的分析器的数量
}

//创建分析器池
func NewAnalyzerPool(total uint32, gen GenAnalyzer) (AnalyzerPool, error) {
	etype := reflect.TypeOf(gen())
	genEntity := func() mdw.Entity {
		return gen()
	}

	pool, err := mdw.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}
	dlpool := &myAnalyzerPool{pool:pool, etype:etype}
	return dlpool, nil
}



type myAnalyzerPool struct {
	pool mdw.Pool	//分析器池
	etype reflect.Type	//类型
}

func (spdPool *myAnalyzerPool) Take() (analyzer Analyzer, err error) {
	entity, err := spdPool.pool.Take()
	if err != nil {
		return nil, err
	}

	spd, ok := entity.(Analyzer)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", spdPool.etype)
		panic(errors.New(errMsg))
	}

	return spd, nil
}

func (spdPool *myAnalyzerPool) Return(analyzer Analyzer) error {
	return spdPool.pool.Return(analyzer)
}

func (spdPool *myAnalyzerPool) Total() uint32 {
	return spdPool.pool.Total()
}

func (spdPool *myAnalyzerPool) Used() uint32 {
	return spdPool.pool.Used()
}
