package client

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/task"
)

// ListTask lists tasks in the given namespace and group
func (c APIClient) ListTask(service string, namespace, group string, cb func(*task.TaskInfo) error) (retErr error) {
	ctx, cancel := context.WithCancel(c.Ctx())
	defer cancel()
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()

	req := &task.ListTaskRequest{
		Group: &task.Group{
			Namespace: namespace,
			Group:     group,
		},
	}
	var stream interface {
		Recv() (*task.TaskInfo, error)
	}
	var err error
	switch service {
	case "pps":
		stream, err = c.PpsAPIClient.ListTask(ctx, req)
	case "pfs":
		stream, err = c.PfsAPIClient.ListTask(ctx, req)
	default:
		return errors.Errorf("%s is not a valid task service", service)
	}
	if err != nil {
		return errors.EnsureStack(err)
	}
	for {
		ti, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(ti); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}
