package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"google.golang.org/protobuf/types/known/anypb"
)

type cache struct {
	*apiServer
	tag string
}

func (a *apiServer) newCache(tag string) *cache {
	return &cache{
		apiServer: a,
		tag:       tag,
	}
}

func (c *cache) Get(ctx context.Context, key string) (*anypb.Any, error) {
	output, err := c.getCache(ctx, key)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (c *cache) Put(ctx context.Context, key string, output *anypb.Any) error {
	var handleStrs []string
	switch {
	case output.MessageIs(&ShardTaskResult{}):
	case output.MessageIs(&CompactTaskResult{}):
		ct, err := deserializeCompactTaskResult(output)
		if err != nil {
			return err
		}
		handleStrs = append(handleStrs, ct.Handle)
	case output.MessageIs(&ConcatTaskResult{}):
		ct, err := deserializeConcatTaskResult(output)
		if err != nil {
			return err
		}
		handleStrs = append(handleStrs, ct.Handle)
	case output.MessageIs(&ValidateTaskResult{}):
	default:
		return errors.Errorf("unrecognized any type (%v) in compaction cache", output.TypeUrl)
	}
	var handles []*fileset.Handle
	for _, handleStr := range handleStrs {
		handle, err := fileset.ParseHandle(handleStr)
		if err != nil {
			return err
		}
		handles = append(handles, handle)
	}
	return c.putCache(ctx, key, output, handles, c.tag)
}

func (c *cache) clear(ctx context.Context) error {
	return c.clearCache(ctx, c.tag)
}

func (a *apiServer) putCache(ctx context.Context, key string, value *anypb.Any, handles []*fileset.Handle, tag string) error {
	return a.cache.Put(ctx, key, value, handles, tag)
}

func (a *apiServer) getCache(ctx context.Context, key string) (*anypb.Any, error) {
	return a.cache.Get(ctx, key)
}

func (a *apiServer) clearCache(ctx context.Context, tagPrefix string) error {
	return a.cache.Clear(ctx, tagPrefix)
}
