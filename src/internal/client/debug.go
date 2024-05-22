package client

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
)

// Profile collects a set of pprof profiles.
func (c APIClient) Profile(ctx context.Context, profile *debug.Profile, filter *debug.Filter, w io.Writer) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	profileC, err := c.DebugClient.Profile(ctx, &debug.ProfileRequest{
		Profile: profile,
		Filter:  filter,
	})
	if err != nil {
		return err
	}
	return grpcutil.WriteFromStreamingBytesClient(profileC, w)
}

// Binary collects a set of binaries.
func (c APIClient) Binary(filter *debug.Filter, w io.Writer) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	binaryC, err := c.DebugClient.Binary(c.Ctx(), &debug.BinaryRequest{Filter: filter})
	if err != nil {
		return err
	}
	return grpcutil.WriteFromStreamingBytesClient(binaryC, w)
}

// Dump collects a standard set of debugging information.
func (c APIClient) Dump(filter *debug.Filter, limit int64, w io.Writer) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	ctx, cf := context.WithCancel(c.Ctx())
	defer cf()
	tpl, err := c.DebugClient.GetDumpV2Template(ctx, &debug.GetDumpV2TemplateRequest{})
	if err != nil {
		return err
	}
	dumpC, err := c.DebugClient.DumpV2(ctx, tpl.Request)
	if err != nil {
		return err
	}
	var d *debug.DumpChunk
	for d, err = dumpC.Recv(); !errors.Is(err, io.EOF); d, err = dumpC.Recv() {
		if err != nil {
			return err
		}
		if content := d.GetContent(); content != nil {
			if _, err = w.Write(content.Content); err != nil {
				return errors.Wrap(err, "write dump contents")
			}
		}
	}
	return nil
}
