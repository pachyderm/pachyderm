package client

import (
	"github.com/pachyderm/pachyderm/v2/src/catalog"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
)

// Query queries the data catalog
func (c APIClient) Query(q string) ([]string, error) {
	resp, err := c.CatalogClient.Query(
		c.Ctx(),
		&catalog.QueryRequest{
			Query: q,
		},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return resp.Results, nil
}
