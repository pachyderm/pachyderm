package kv

import (
	"context"
)

// ValueCallback is the type of functions used to access values
type ValueCallback = func([]byte) error

// GetPut supports the basic Get and Put operations
type GetPut interface {
	// Get looks up the value that corresponds to key and calls cb once if the key exists
	// if the key does not exist then a pacherr.ErrNotExist is returned
	Get(ctx context.Context, key []byte, cb ValueCallback) error
	// Put creates an entry mapping key to value, overwriting any previous mapping.
	Put(ctx context.Context, key, value []byte) error
}

// Store is a key-value store
type Store interface {
	GetPut
	Delete(ctx context.Context, key []byte) error
	Exists(ctx context.Context, key []byte) (bool, error)
	Walk(ctx context.Context, prefix []byte, cb func(key []byte) error) error
}

// KeyAfter returns the key immediately after x
func KeyAfter(x []byte) []byte {
	ret := append([]byte{}, x...)
	ret = append(ret, 0x00)
	return ret
}
