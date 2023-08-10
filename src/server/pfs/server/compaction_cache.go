package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"google.golang.org/protobuf/types/known/anypb"
)

type cache struct {
	driver *driver
	tag    string
}

func (d *driver) newCache(tag string) *cache {
	return &cache{
		driver: d,
		tag:    tag,
	}
}

func (c *cache) Get(ctx context.Context, key string) (*anypb.Any, error) {
	output, err := c.driver.getCache(ctx, key)
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
		handleStrs = append(handleStrs, ct.Id)
	case output.MessageIs(&ConcatTaskResult{}):
		ct, err := deserializeConcatTaskResult(output)
		if err != nil {
			return err
		}
		handleStrs = append(handleStrs, ct.Id)
	case output.MessageIs(&ValidateTaskResult{}):
	default:
		return errors.Errorf("unrecognized any type (%v) in compaction cache", output.TypeUrl)
	}
	hs, err := fileset.ParseHandles(handleStrs)
	if err != nil {
		return err
	}
	return c.driver.putCache(ctx, key, output, hs, c.tag)
}

func (c *cache) clear(ctx context.Context) error {
	return c.driver.clearCache(ctx, c.tag)
}
