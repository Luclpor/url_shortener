package pool

import (
	"errors"
	"sync"
)

// ErrNilConstructor is returned when New receives a nil constructor.
var ErrNilConstructor = errors.New("pool constructor is nil")

// Resetter describes values that can clear their internal state.
type Resetter interface {
	Reset()
}

// Pool stores reusable objects of one resettable type.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New creates a pool that uses newFn to allocate objects when the pool is empty.
func New[T Resetter](newFn func() T) (*Pool[T], error) {
	if newFn == nil {
		return nil, ErrNilConstructor
	}

	p := &Pool[T]{}
	p.pool.New = func() any {
		return newFn()
	}
	return p, nil
}

// Get returns an object from the pool.
func (p *Pool[T]) Get() T {
	obj := p.pool.Get()
	if obj == nil {
		var zero T
		return zero
	}
	return obj.(T)
}

// Put resets obj and returns it to the pool.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
