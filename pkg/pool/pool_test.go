package pool

import "testing"

type resettableBuffer struct {
	values     []int
	resetCalls int
}

func (b *resettableBuffer) Reset() {
	b.values = b.values[:0]
	b.resetCalls++
}

func TestPoolGetUsesConstructor(t *testing.T) {
	created := 0
	p := New(func() *resettableBuffer {
		created++
		return &resettableBuffer{values: make([]int, 0, 4)}
	})

	got := p.Get()

	if got == nil {
		t.Fatal("Get returned nil")
	}
	if created != 1 {
		t.Fatalf("constructor was called %d times, want 1", created)
	}
	if cap(got.values) != 4 {
		t.Fatalf("unexpected buffer capacity: %d", cap(got.values))
	}
}

func TestPoolPutResetsObject(t *testing.T) {
	p := New(func() *resettableBuffer {
		return &resettableBuffer{values: make([]int, 0, 8)}
	})
	buf := p.Get()
	buf.values = append(buf.values, 1, 2, 3)

	p.Put(buf)

	if len(buf.values) != 0 {
		t.Fatalf("Put did not reset object values: %v", buf.values)
	}
	if cap(buf.values) != 8 {
		t.Fatalf("Put should keep reusable capacity, got %d", cap(buf.values))
	}
	if buf.resetCalls != 1 {
		t.Fatalf("Reset was called %d times, want 1", buf.resetCalls)
	}
}

func TestPoolWorksWithoutConstructor(t *testing.T) {
	p := New[*resettableBuffer]()
	buf := &resettableBuffer{values: []int{1, 2, 3}}

	p.Put(buf)
	got := p.Get()

	if got != buf {
		t.Fatalf("Get returned %p, want %p", got, buf)
	}
	if len(got.values) != 0 {
		t.Fatalf("stored object was not reset: %v", got.values)
	}
}
