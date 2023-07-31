package queux

import (
	"container/list"
	"context"
	"errors"
	"sync"
)

var (
	QueuxClosed = errors.New("queux.closed")
)

type QueuxMemory[T any] struct {
	locker sync.Mutex
	list   *list.List
	read   chan struct{}
}

func NewQueuxMemory[T any]() *QueuxMemory[T] {
	return &QueuxMemory[T]{
		list: list.New(),
		read: make(chan struct{}),
	}
}

func (impl *QueuxMemory[T]) Push(v T) error {
	impl.locker.Lock()
	defer impl.locker.Unlock()
	impl.list.PushBack(v)
	select {
	case impl.read <- struct{}{}:
		return nil
	default:
		return nil
	}
}

func (impl *QueuxMemory[T]) Get() (T, bool) {
	impl.locker.Lock()
	defer impl.locker.Unlock()
	if impl.list.Len() > 0 {
		el := impl.list.Front()
		impl.list.Remove(el)
		return el.Value.(T), true
	}
	var noop T
	return noop, false
}

func (impl *QueuxMemory[T]) GetOrWait(ctx context.Context) (T, error) {
	var noop T
	if v, ok := impl.Get(); ok {
		return v, nil
	}
	for {
		select {
		case _, ok := <-impl.read:
			if !ok {
				return noop, QueuxClosed
			}
			if v, ok := impl.Get(); ok {
				return v, nil
			}
			continue
		case <-ctx.Done():
			return noop, ctx.Err()
		}
	}
}
