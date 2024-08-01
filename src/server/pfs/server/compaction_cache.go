package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/server/driver"
	"google.golang.org/protobuf/types/known/anypb"
)

type Cache struct {
	driver *driver.Driver
	tag    string
}

func NewCache(tag string, d *driver.Driver) *Cache {
	return &Cache{
		driver: d,
		tag:    tag,
	}
}

func (c *Cache) Get(ctx context.Context, key string) (*anypb.Any, error) {
	output, err := c.driver.GetCache(ctx, key)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Cache) Put(ctx context.Context, key string, output *anypb.Any) error {
	var fileSetIds []string
	switch {
	case output.MessageIs(&ShardTaskResult{}):
	case output.MessageIs(&CompactTaskResult{}):
		ct, err := deserializeCompactTaskResult(output)
		if err != nil {
			return err
		}
		fileSetIds = append(fileSetIds, ct.Id)
	case output.MessageIs(&ConcatTaskResult{}):
		ct, err := deserializeConcatTaskResult(output)
		if err != nil {
			return err
		}
		fileSetIds = append(fileSetIds, ct.Id)
	case output.MessageIs(&ValidateTaskResult{}):
	default:
		return errors.Errorf("unrecognized any type (%v) in compaction cache", output.TypeUrl)
	}
	var fsids []fileset.ID
	for _, id := range fileSetIds {
		fsid, err := fileset.ParseID(id)
		if err != nil {
			return err
		}
		fsids = append(fsids, *fsid)
	}
	return c.driver.PutCache(ctx, key, output, fsids, c.tag)
}

func (c *Cache) Clear(ctx context.Context) error {
	return c.driver.ClearCache(ctx, c.tag)
}
