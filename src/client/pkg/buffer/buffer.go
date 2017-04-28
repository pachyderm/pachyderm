// Package buffer exports a global buffer pool.  Clients who reguarly
// need access to buffers are expected to use this package as opposed to
// allocating new buffers themselves, in order to reduce overall memory
// allocation.
package buffer

import (
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		// We use 2MB as the default buffer size because this buffer pool
		// is typically used for gRPC and 2MB is:
		// 1. Reasonably smaller than the max gRPC size (which is 20MB)
		// 2. Small enough that having hundreds of these buffers won't
		// overwhelm the node
		// 3. Large enough for message-sending to be efficient
		return make([]byte, 2*1024*1024)
	},
}

// Get returns a buffer.  The buffer may or may not be freshly allocated.
func Get() []byte {
	return bufPool.Get().([]byte)
}

// Put returns the buffer to the pool.
func Put(buf []byte) {
	bufPool.Put(buf)
}
