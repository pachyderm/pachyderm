package task

import (
	"context"

	"github.com/gogo/protobuf/types"
)

// TODO: Document when we have committed to the interface.

type Service interface {
	// TODO: Should group ID be a parameter?
	Doer(ctx context.Context, namespace, groupID string, cb func(context.Context, Doer) error) error
	Maker(namespace string) Maker
	// TaskCount returns how many tasks are in a namespace and how many are claimed.
	TaskCount(ctx context.Context, namespace string) (tasks int64, claims int64, err error)
}

type Doer interface {
	Do(ctx context.Context, any *types.Any) (*types.Any, error)
	DoBatch(ctx context.Context, anys []*types.Any, cb CollectFunc) error
	DoChan(ctx context.Context, anyChan chan *types.Any, cb CollectFunc) error
}

type Maker interface {
	Make(ctx context.Context, cb ProcessFunc) error
}

type CollectFunc = func(int64, *types.Any, error) error
type ProcessFunc = func(context.Context, *types.Any) (*types.Any, error)
