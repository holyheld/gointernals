package pool

import (
	"bytes"
	"fmt"
	"sync"
)

const Unsized = 0

const (
	growthFactor       = 2
	maxUnsizedPoolSize = 64 * 1024
)

type BufferPool struct {
	size        int64
	maxCapacity int64
	p           sync.Pool
}

// NewBufferPool creates a new buffer pool with the given default size.
// If defaultSize is 0, the pool will create a new empty buffer for each Pool.New call
// Else, the pool will create a new buffer with the given cap for each Pool.New call
//
// Panics if defaultSize is negative.
func NewBufferPool(defaultSize int64) *BufferPool {
	if defaultSize < 0 {
		panic(fmt.Errorf("defaultSize is negative: %d", defaultSize))
	}

	maxCap := defaultSize * growthFactor
	if defaultSize == 0 {
		maxCap = maxUnsizedPoolSize
	}

	return &BufferPool{
		size:        defaultSize,
		maxCapacity: maxCap,
		p: sync.Pool{
			New: func() any {
				if defaultSize == 0 {
					return new(bytes.Buffer)
				}

				return bytes.NewBuffer(make([]byte, 0, defaultSize))
			},
		},
	}
}

// Get returns a buffer from the pool.
func (p *BufferPool) Get() *bytes.Buffer {
	//nolint:forcetypeassert // pool contains only entries of *bytes.Buffer
	v := p.p.Get().(*bytes.Buffer)
	v.Reset()

	return v
}

// Put resets the buffer and puts it back to the pool.
func (p *BufferPool) Put(v *bytes.Buffer) {
	if int64(v.Cap()) > p.maxCapacity {
		// drop it to be collected by GC
		return
	}

	p.p.Put(v)
}

func (p *BufferPool) Size() int64 {
	return p.size
}
