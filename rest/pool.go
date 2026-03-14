package rest

import (
	"github.com/holyheld/pool"
)

const (
	kilobyte                          = 1 << 10
	defaultTransferEncodingBufferSize = 4 * kilobyte
)

var (
	bufferPoolUnsized = pool.NewBufferPool(pool.Unsized)
	bufferPool2k      = pool.NewBufferPool(2 * kilobyte)
	bufferPool4k      = pool.NewBufferPool(4 * kilobyte)
	bufferPool8k      = pool.NewBufferPool(8 * kilobyte)
	bufferPool32k     = pool.NewBufferPool(32 * kilobyte)
	bufferPool64k     = pool.NewBufferPool(64 * kilobyte)
	bufferPool128k    = pool.NewBufferPool(128 * kilobyte)
	bufferPool1M      = pool.NewBufferPool(1024 * kilobyte)
	bufferPool16M     = pool.NewBufferPool(16 * 1024 * kilobyte)
)

// getPool returns pool selected by size. If size is 0 returns unsized pool.
func getPool(size int64) *pool.BufferPool {
	switch {
	case size == bufferPoolUnsized.Size():
		return bufferPoolUnsized
	case size < bufferPool2k.Size():
		return bufferPool2k
	case size < bufferPool4k.Size():
		return bufferPool4k
	case size < bufferPool8k.Size():
		return bufferPool8k
	case size < bufferPool32k.Size():
		return bufferPool32k
	case size < bufferPool64k.Size():
		return bufferPool64k
	case size < bufferPool128k.Size():
		return bufferPool128k
	case size < bufferPool1M.Size():
		return bufferPool1M
	case size < bufferPool16M.Size():
		return bufferPool16M
	default:
		return bufferPoolUnsized
	}
}
