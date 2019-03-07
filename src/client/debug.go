package client

import (
	"io"
	"time"

	"github.com/gogo/protobuf/types"
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

func (c APIClient) Profile(profile string, duration time.Duration, w io.Writer) error {
	var d *types.Duration
	if duration != 0 {
		d = types.DurationProto(duration)
	}
	profileClient, err := c.DebugClient.Profile(c.Ctx(), &debug.ProfileRequest{
		Profile:  profile,
		Duration: d,
	})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return grpcutil.ScrubGRPC(grpcutil.WriteFromStreamingBytesClient(profileClient, w))
}

func (c APIClient) Binary(w io.Writer) error {
	binaryClient, err := c.DebugClient.Binary(c.Ctx(), &debug.BinaryRequest{})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return grpcutil.ScrubGRPC(grpcutil.WriteFromStreamingBytesClient(binaryClient, w))
}
