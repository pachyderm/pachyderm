package kv

import (
	"context"
	"sync"
)

// Pool manages a pool of buffers, all the same size.
//
// The pointer to slice stuff is to make the linter happy.
type Pool struct {
	maxBufferSize int
	pool          sync.Pool
}

func NewPool(maxBufferSize int) *Pool {
	return &Pool{
		maxBufferSize: maxBufferSize,
		pool: sync.Pool{
			New: func() any {
				buf := make([]byte, maxBufferSize)
				return &buf
			},
		},
	}
}

func (p *Pool) GetF(ctx context.Context, s Getter, key []byte, cb ValueCallback) error {
	buf := p.acquire()
	defer p.release(buf)
	n, err := s.Get(ctx, key, *buf)
	if err != nil {
		return err
	}
	return cb((*buf)[:n])
}

func (p *Pool) acquire() *[]byte {
	return p.pool.Get().(*[]byte)
}

func (p *Pool) release(x *[]byte) {
	p.pool.Put(x)
}
