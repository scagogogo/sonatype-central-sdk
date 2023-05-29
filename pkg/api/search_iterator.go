package api

import (
	"context"
	"errors"
	"github.com/golang-infrastructure/go-iterator"
	"github.com/golang-infrastructure/go-queue"
	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
)

var (
	// ErrQueryIteratorEnd 迭代器已经遍历完毕
	ErrQueryIteratorEnd = errors.New("query iterator ended")
)

// SearchIterator 把搜索结果以一个迭代器的形式返回
type SearchIterator[Doc any] struct {

	// 搜索参数
	search *request.SearchRequest

	// 缓存大小
	buff *queue.LinkedQueue[Doc]

	// 记录遍历位置
	total     int
	current   int
	nextStart int

	// 迭代器是否发生错误
	err error
}

var _ iterator.ErrorableIterator[any] = &SearchIterator[any]{}

func NewSearchIterator[Doc any](search *request.SearchRequest) *SearchIterator[Doc] {
	return &SearchIterator[Doc]{
		search:    search,
		buff:      queue.NewLinkedQueue[Doc](),
		total:     -1,
		current:   0,
		nextStart: 0,
		err:       nil,
	}
}

func (x *SearchIterator[Doc]) ToSlice() ([]Doc, error) {
	slice := make([]Doc, 0)
	for {
		hasNext, err := x.NextE()
		if err != nil {
			return slice, err
		}
		if !hasNext {
			return slice, nil
		}
		artifact, err := x.ValueE()
		if err != nil {
			return slice, err
		}
		slice = append(slice, artifact)
	}
}

func (x *SearchIterator[Doc]) Next() bool {
	hasNext, _ := x.NextE()
	return hasNext
}

func (x *SearchIterator[Doc]) Value() Doc {
	value, _ := x.ValueE()
	return value
}

func (x *SearchIterator[Doc]) NextE() (bool, error) {

	if x.err != nil {
		return false, x.err
	}

	if x.total < 0 {
		// 初始化
		r, err := SearchRequest[Doc](context.Background(), x.search)
		if err != nil {
			x.err = err
			return false, err
		}
		x.total = r.ResponseBody.NumFound
		x.nextStart = x.nextStart + len(r.ResponseBody.Docs)
		err = x.buff.Put(r.ResponseBody.Docs...)
		if err != nil {
			x.err = err
			return false, err
		}
		return x.current < x.total, nil
	} else {
		// 已经初始化过了，判断是否需要拿新的数据
		if x.buff.IsNotEmpty() {
			return true, nil
		}
		if x.current >= x.total {
			return false, nil
		}
		x.search.Start = x.nextStart
		r, err := SearchRequest[Doc](context.Background(), x.search)
		if err != nil {
			x.err = err
			return false, err
		}
		x.nextStart = x.nextStart + len(r.ResponseBody.Docs)
		err = x.buff.Put(r.ResponseBody.Docs...)
		if err != nil {
			x.err = err
			return false, err
		}
		return x.current < x.total, nil
	}
}

func (x *SearchIterator[Doc]) ValueE() (Doc, error) {
	if x.buff.IsEmpty() {
		var zero Doc
		return zero, ErrQueryIteratorEnd
	}
	x.current++
	return x.buff.TakeE()
}
