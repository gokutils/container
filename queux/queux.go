package queux

import (
	"context"
	"sync"

	"github.com/gokutils/container/list"
	lsync "github.com/gokutils/container/sync"
)

type QueuxMemory[T any] struct {
	locker *sync.Mutex
	list   *list.List[T]
	cond   *lsync.Cond
}

func NewQueuxMemory[T any]() *QueuxMemory[T] {
	locker := &sync.Mutex{}
	return &QueuxMemory[T]{
		locker: locker,
		list:   list.New[T](),
		cond:   lsync.NewCond(locker),
	}
}

func (impl *QueuxMemory[T]) Push(v T) error {
	impl.locker.Lock()
	impl.list.PushBack(v)
	impl.locker.Unlock()
	impl.cond.Signal()
	return nil
}

func (impl *QueuxMemory[T]) get() (T, bool) {
	if impl.list.Len() > 0 {
		el := impl.list.PopFront()
		return el, true
	}
	var noop T
	return noop, false
}

func (impl *QueuxMemory[T]) Get() (T, bool) {
	impl.locker.Lock()
	defer impl.locker.Unlock()
	return impl.get()
}

func (impl *QueuxMemory[T]) GetOrWait(ctx context.Context) (T, error) {
	var noop T
	impl.locker.Lock()
	defer impl.locker.Unlock()
	if v, ok := impl.get(); ok {
		return v, nil
	}

	for {
		err := impl.cond.Wait(ctx)
		if err != nil {
			return noop, err
		}
		if v, ok := impl.get(); ok {
			return v, nil
		}
	}
}
