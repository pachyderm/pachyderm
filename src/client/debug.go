package client

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

func (c APIClient) Goro(w io.Writer) error {
	goroClient, err := c.DebugClient.Goro(c.Ctx(), &debug.GoroRequest{})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return grpcutil.ScrubGRPC(grpcutil.WriteFromStreamingBytesClient(goroClient, w))
}
