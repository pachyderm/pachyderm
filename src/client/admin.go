package client

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

// Extract all cluster state as a series of operations.
func (c APIClient) Extract() ([]*admin.Op, error) {
	extractClient, err := c.AdminAPIClient.Extract(c.Ctx(), &admin.ExtractRequest{})
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	var result []*admin.Op
	for {
		op, err := extractClient.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, grpcutil.ScrubGRPC(err)
		}
		result = append(result, op)
	}
	return result, nil
}

// Restore cluster state from an extract series of operations.
func (c APIClient) Restore(ops []*admin.Op) (retErr error) {
	restoreClient, err := c.AdminAPIClient.Restore(c.Ctx())
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	defer func() {
		if _, err := restoreClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = grpcutil.ScrubGRPC(err)
		}
	}()
	for _, op := range ops {
		if err := restoreClient.Send(&admin.RestoreRequest{Op: op}); err != nil {
			return grpcutil.ScrubGRPC(err)
		}
	}
	return nil
}
