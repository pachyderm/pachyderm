package buffer

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, grpcutil.MaxMsgSize/10)
	},
}

func Get() []byte {
	return bufPool.Get().([]byte)
}

func Put(buf []byte) {
	bufPool.Put(buf)
}
