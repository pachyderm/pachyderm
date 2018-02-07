package client

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
)

// Extract all cluster state, call f with each operation.
func (c APIClient) Extract(f func(op *admin.Op) error) error {
	extractClient, err := c.AdminAPIClient.Extract(c.Ctx(), &admin.ExtractRequest{})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	for {
		op, err := extractClient.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return grpcutil.ScrubGRPC(err)
		}
		if err := f(op); err != nil {
			return err
		}
	}
	return nil
}

// ExtractAll cluster state as a slice of operations.
func (c APIClient) ExtractAll() ([]*admin.Op, error) {
	var result []*admin.Op
	if err := c.Extract(func(op *admin.Op) error {
		result = append(result, op)
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// ExtractWriter extracts all cluster state and marshals it to w.
func (c APIClient) ExtractWriter(w io.Writer) error {
	writer := pbutil.NewWriter(w)
	return c.Extract(func(op *admin.Op) error {
		return writer.Write(op)
	})
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

// RestoreReader restores cluster state from a reader containing marshaled ops.
// Such as those written by ExtractWriter.
func (c APIClient) RestoreReader(r io.Reader) (retErr error) {
	restoreClient, err := c.AdminAPIClient.Restore(c.Ctx())
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	defer func() {
		if _, err := restoreClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = grpcutil.ScrubGRPC(err)
		}
	}()
	reader := pbutil.NewReader(r)
	op := &admin.Op{}
	for {
		if err := reader.Read(op); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := restoreClient.Send(&admin.RestoreRequest{Op: op}); err != nil {
			return grpcutil.ScrubGRPC(err)
		}
	}
	return nil
}

// RestoreFrom restores state from another cluster which can be access through otherC.
func (c APIClient) RestoreFrom(otherC *APIClient) (retErr error) {
	restoreClient, err := c.AdminAPIClient.Restore(c.Ctx())
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	defer func() {
		if _, err := restoreClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = grpcutil.ScrubGRPC(err)
		}
	}()
	return otherC.Extract(func(op *admin.Op) error {
		return restoreClient.Send(&admin.RestoreRequest{Op: op})
	})
}
