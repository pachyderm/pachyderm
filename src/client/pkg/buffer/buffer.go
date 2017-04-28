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

func Get() []byte {
	return bufPool.Get().([]byte)
}

func Put(buf []byte) {
	bufPool.Put(buf)
}
