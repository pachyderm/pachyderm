package client

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/taskapi"
)

// ListTask lists tasks in the given namespace and group
func (c APIClient) ListTask(namespace string) ([]*taskapi.TaskInfo, error) {
	ctx, cancel := context.WithCancel(c.Ctx())
	defer cancel()

	stream, err := c.TaskClient.ListTask(
		ctx,
		&taskapi.ListTaskRequest{
			Namespace: namespace,
		},
	)
	var out []*taskapi.TaskInfo
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	for {
		ti, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, grpcutil.ScrubGRPC(err)
		}
		out = append(out, ti)
	}
	return out, grpcutil.ScrubGRPC(err)
}
