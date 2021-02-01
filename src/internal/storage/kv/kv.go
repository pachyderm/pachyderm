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

// GetPut supports the basic GetF and Put operations
type GetPut interface {
	Put(ctx context.Context, key, value []byte) error
	GetF(ctx context.Context, key []byte, cb ValueCallback) error
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
func AssertNotModified(x []byte, cb ValueCallback) error {
	before := crc64.Checksum(x, crc64Tab)
	err := cb(x)
	after := crc64.Checksum(x, crc64Tab)
	if before != after {
		panic("buffer was modified")
	}
	return err
}
