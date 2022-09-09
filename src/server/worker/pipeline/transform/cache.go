package transform

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
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
	switch {
	case datum.IsTaskResult(output):
		var err error
		fileSetIds, err = datum.TaskResultFileSets(output)
		if err != nil {
			return err
		}
	case types.Is(output, &CreateParallelDatumsTaskResult{}):
		cpdt, err := deserializeCreateParallelDatumsTaskResult(output)
		if err != nil {
			return err
		}
		fileSetIds = append(fileSetIds, cpdt.FileSetId)
	case types.Is(output, &CreateSerialDatumsTaskResult{}):
		csdt, err := deserializeCreateSerialDatumsTaskResult(output)
		if err != nil {
			return err
		}
		fileSetIds = append(fileSetIds, csdt.FileSetId, csdt.OutputDeleteFileSetId, csdt.MetaDeleteFileSetId)
	case types.Is(output, &CreateDatumSetsTaskResult{}):
	case types.Is(output, &DatumSetTaskResult{}):
		dst, err := deserializeDatumSetTaskResult(output)
		if err != nil {
			return err
		}
		fileSetIds = append(fileSetIds, dst.OutputFileSetId, dst.MetaFileSetId)
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
