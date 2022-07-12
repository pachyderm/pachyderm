package transform

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

type cache struct {
	pachClient *client.APIClient
	tag        string
}

func newCache(pachClient *client.APIClient, tag string) *cache {
	return &cache{
		pachClient: pachClient,
		tag:        tag,
	}
}

func (c *cache) Get(ctx context.Context, key string) (*types.Any, error) {
	resp, err := c.pachClient.PfsAPIClient.GetCache(ctx, &pfs.GetCacheRequest{Key: key})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return resp.Value, nil
}

func (c *cache) Put(ctx context.Context, key string, output *types.Any) error {
	var fileSetIds []string
	// TODO: Add caching for datum tasks.
	switch {
	case types.Is(output, &ComputeParallelDatumsTaskResult{}):
		cpdt, err := deserializeComputeParallelDatumsTaskResult(output)
		if err != nil {
			return err
		}
		fileSetIds = append(fileSetIds, cpdt.FileSetId)
	case types.Is(output, &ComputeSerialDatumsTaskResult{}):
		csdt, err := deserializeComputeSerialDatumsTaskResult(output)
		if err != nil {
			return err
		}
		fileSetIds = append(fileSetIds, csdt.FileSetId, csdt.OutputDeleteFileSetId, csdt.MetaDeleteFileSetId)
	case types.Is(output, &DatumSet{}):
		ds, err := deserializeDatumSet(output)
		if err != nil {
			return err
		}
		fileSetIds = append(fileSetIds, ds.FileSetId, ds.OutputFileSetId, ds.MetaFileSetId)
	default:
		return errors.Errorf("unrecognized any type (%v) in transform cache", output.TypeUrl)
	}
	_, err := c.pachClient.PfsAPIClient.PutCache(ctx, &pfs.PutCacheRequest{
		Key:        key,
		Value:      output,
		FileSetIds: fileSetIds,
		Tag:        c.tag,
	})
	return errors.EnsureStack(err)
}

func (c *cache) clear(ctx context.Context) error {
	_, err := c.pachClient.PfsAPIClient.ClearCache(ctx, &pfs.ClearCacheRequest{
		TagPrefix: c.tag,
	})
	return errors.EnsureStack(err)
}
