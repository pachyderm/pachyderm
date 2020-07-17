package client

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

// Profile writes a pprof profile for pachd to w.
func (c APIClient) Profile(profile *debug.Profile, worker *debug.Worker, w io.Writer) error {
	profileClient, err := c.DebugClient.Profile(c.Ctx(), &debug.ProfileRequest{
		Profile: profile,
		Worker:  worker,
	})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return grpcutil.ScrubGRPC(grpcutil.WriteFromStreamingBytesClient(profileClient, w))
}

// Binary writes the running pachd binary to w.
func (c APIClient) Binary(worker *debug.Worker, w io.Writer) error {
	binaryClient, err := c.DebugClient.Binary(c.Ctx(), &debug.BinaryRequest{
		Worker: worker,
	})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return grpcutil.ScrubGRPC(grpcutil.WriteFromStreamingBytesClient(binaryClient, w))
}

func (c APIClient) SOS(worker *debug.Worker, w io.Writer) error {
	sosClient, err := c.DebugClient.SOS(c.Ctx(), &debug.SOSRequest{
		Worker: worker,
	})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return grpcutil.ScrubGRPC(grpcutil.WriteFromStreamingBytesClient(sosClient, w))
}
