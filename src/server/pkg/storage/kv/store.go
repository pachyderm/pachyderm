package kv

import "context"

type Store interface {
	Put(ctx context.Context, key string, value []byte) error
	GetF(ctx context.Context, key string, cb func(data []byte) error) error
	Exists(ctx context.Context, key string) bool
	Delete(ctx context.Context, key string) error
	Walk(ctx context.Context, prefix string, cb func(key string) error) error

	IsNotExist(error) bool
}
