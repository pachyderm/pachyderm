package client

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

// Dump returns debug information from the server.
func (c APIClient) Dump(w io.Writer) error {
	goroClient, err := c.DebugClient.Dump(c.Ctx(), &debug.DumpRequest{})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return grpcutil.ScrubGRPC(grpcutil.WriteFromStreamingBytesClient(goroClient, w))
}
