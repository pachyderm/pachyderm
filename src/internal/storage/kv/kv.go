package kv

import (
	"context"
	"errors"
	"hash/crc64"
)

// ErrKeyNotFound is returned when the store does not contain a key
var ErrKeyNotFound = errors.New("key not found")

// ValueCallback is the type of functions used to access values
type ValueCallback = func([]byte) error

// GetPut supports the basic Get and Put operations
type GetPut interface {
	// Get lookups up the value that corresponds to key and calls cb once if the key exists
	// if the key does not exist ErrKeyNotFound is returned
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

var crc64Tab = crc64.MakeTable(crc64.ISO)

// AssertNotModified calls cb with x, but panics if x changed during the execution of cb
// TODO: remove calls to this before 2.0 GA
func AssertNotModified(x []byte, cb ValueCallback) error {
	before := crc64.Checksum(x, crc64Tab)
	err := cb(x)
	after := crc64.Checksum(x, crc64Tab)
	if before != after {
		panic("buffer was modified")
	}
	return err
}
