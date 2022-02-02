package client

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/task"
)

// ListTask lists tasks in the given namespace and group
func (c APIClient) ListTask(service string, namespace, group string) (infos []*task.TaskInfo, retErr error) {
	ctx, cancel := context.WithCancel(c.Ctx())
	defer cancel()
	defer func() {
		if retErr != nil {
			retErr = grpcutil.ScrubGRPC(retErr)
		}
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
		return nil, errors.Errorf("%s is not a valid task service", service)
	}
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	var out []*task.TaskInfo
	for {
		ti, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, errors.EnsureStack(err)
		}
		out = append(out, ti)
	}
	return out, nil
}
