package kv

import (
	"context"
	"sync"
)

type Pool struct {
	pool sync.Pool
}

func NewPool(maxBufferSize int) *Pool {
	return &Pool{
		pool: sync.Pool{
			New: func() any {
				return make([]byte, maxBufferSize)
			},
		},
	}
}

func (p *Pool) GetF(ctx context.Context, s Getter, key []byte, cb ValueCallback) error {
	buf := p.acquire()
	defer p.release(buf)
	n, err := s.Get(ctx, key, buf)
	if err != nil {
		return err
	}
	return cb(buf[:n])
}

func (p *Pool) acquire() []byte {
	return p.pool.Get().([]byte)
}

func (p *Pool) release(x []byte) {
	p.pool.Put(x)
}
