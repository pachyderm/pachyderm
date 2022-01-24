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
}

func newCache(pachClient *client.APIClient) *cache {
	return &cache{pachClient: pachClient}
}

func (c *cache) Get(ctx context.Context, key string) (*types.Any, error) {
	resp, err := c.pachClient.PfsAPIClient.GetCache(ctx, &pfs.GetCacheRequest{Key: key})
	if err != nil {
		return nil, err
	}
	return resp.Value, nil
}

func (c *cache) Put(ctx context.Context, key string, output *types.Any) error {
	var fileSetIds []string
	switch {
	case types.Is(output, &UploadDatumsTaskResult{}):
		udt, err := deserializeUploadDatumsTaskResult(output)
		if err != nil {
			return err
		}
		fileSetIds = append(fileSetIds, udt.FileSetId)
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
		fileSetIds = append(fileSetIds, csdt.FileSetId, csdt.DeleteFileSetId)
	case types.Is(output, &CreateDatumSetsTaskResult{}):
		cdst, err := deserializeCreateDatumSetsTaskResult(output)
		if err != nil {
			return err
		}
		fileSetIds = append(fileSetIds, cdst.FileSetId)
	case types.Is(output, &DatumSet{}):
		ds, err := deserializeDatumSet(output)
		if err != nil {
			return err
		}
		// TODO: The input file set probably doesn't need to be cached.
		fileSetIds = append(fileSetIds, ds.FileSetId, ds.OutputFileSetId, ds.MetaFileSetId)
	default:
		return errors.Errorf("unrecognized any type (%v) in transform cache", output.TypeUrl)
	}
	_, err := c.pachClient.PfsAPIClient.PutCache(ctx, &pfs.PutCacheRequest{
		Key:        key,
		Value:      output,
		FileSetIds: fileSetIds,
	})
	return err
}
