package grpcutil

import (
	"sync"
)

// BufPool is a wrapper around sync.Pool that makes it a little nicer to use
// for []byte by doing the casting for you and defining the `New` function.
type BufPool struct {
	sync.Pool
}

// NewBufPool creates a new BufPool that returns buffers of the given size.
func NewBufPool(size int) *BufPool {
	return &BufPool{sync.Pool{
		New: func() interface{} { return make([]byte, size) },
	}}
}

// GetBuffer returns a buffer.  The buffer may or may not be freshly
// allocated, and it may or may not be zero-ed.
func (b *BufPool) GetBuffer() []byte {
	return b.Get().([]byte)
}

// PutBuffer returns the buffer to the pool.
func (b *BufPool) PutBuffer(buf []byte) {
	b.Put(buf) //lint:ignore SA6002 []byte is sufficiently pointer-like for our purposes
}

// bufPool is a pool of buffers that are sized for grpc connections
// This buffer size is:
// 1. Reasonably smaller than the max gRPC size
// 2. Small enough that having hundreds of these buffers won't
// overwhelm the node
// 3. Large enough for message-sending to be efficient
var bufPool = NewBufPool(MaxMsgSize / 10)

// GetBuffer returns a buffer.  The buffer may or may not be freshly
// allocated, and it may or may not be zero-ed.
func GetBuffer() []byte {
	return bufPool.GetBuffer()
}

// PutBuffer returns the buffer to the pool.
func PutBuffer(buf []byte) {
	bufPool.PutBuffer(buf)
}
