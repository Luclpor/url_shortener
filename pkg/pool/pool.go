package pool

import "sync"

// Resetter describes values that can clear their internal state.
type Resetter interface {
	Reset()
}

// Pool stores reusable objects of one resettable type.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New creates a pool.
//
// If newFn is provided, it is used to allocate objects when the pool is empty.
func New[T Resetter](newFn ...func() T) *Pool[T] {
	p := &Pool[T]{}
	if len(newFn) > 0 && newFn[0] != nil {
		p.pool.New = func() any {
			return newFn[0]()
		}
	}
	return p
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
