package client

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

// Profile collects a set of pprof profiles.
func (c APIClient) Profile(profile *debug.Profile, filter *debug.Filter, w io.Writer) error {
	profileClient, err := c.DebugClient.Profile(c.Ctx(), &debug.ProfileRequest{
		Profile: profile,
		Filter:  filter,
	})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return grpcutil.ScrubGRPC(grpcutil.WriteFromStreamingBytesClient(profileClient, w))
}

// Binary collects a set of binaries.
func (c APIClient) Binary(filter *debug.Filter, w io.Writer) error {
	binaryClient, err := c.DebugClient.Binary(c.Ctx(), &debug.BinaryRequest{
		Filter: filter,
	})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return grpcutil.ScrubGRPC(grpcutil.WriteFromStreamingBytesClient(binaryClient, w))
}

// Dump collects a standard set of debugging information.
func (c APIClient) Dump(filter *debug.Filter, w io.Writer) error {
	dumpClient, err := c.DebugClient.Dump(c.Ctx(), &debug.DumpRequest{
		Filter: filter,
	})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return grpcutil.ScrubGRPC(grpcutil.WriteFromStreamingBytesClient(dumpClient, w))
}
