package server

import (
	"bufio"
	"context"
	"io"
	"runtime/trace"
	"time"

	"github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (d *debugServer) Trace(req *debug.TraceRequest, s debug.Debug_TraceServer) (retErr error) {
	r, uw := io.Pipe()
	w := bufio.NewWriterSize(uw, units.MiB)
	if err := trace.Start(w); err != nil {
		return status.Error(codes.Unavailable, err.Error())
	}
	go func() {
		var retErr error
		select {
		case <-time.After(req.GetDuration().AsDuration()):
		case <-s.Context().Done():
			errors.JoinInto(&retErr, context.Cause(s.Context()))
		}
		trace.Stop()
		if err := w.Flush(); err != nil {
			errors.JoinInto(&retErr, errors.Wrap(err, "flush"))
		}
		uw.CloseWithError(retErr)
	}()

	p := make([]byte, units.MiB)
	for {
		n, err := r.Read(p)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return errors.Wrap(err, "read")
		}
		if err := s.Send(&debug.TraceChunk{
			Reply: &debug.TraceChunk_Bytes{
				Bytes: wrapperspb.Bytes(p[:n]),
			},
		}); err != nil {
			return errors.Wrap(err, "send")
		}
	}
}
