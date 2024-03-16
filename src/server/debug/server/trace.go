package server

import (
	"bytes"
	"io"
	"time"

	"github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"golang.org/x/exp/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (d *debugServer) Trace(req *debug.TraceRequest, s debug.Debug_TraceServer) (retErr error) {
	fr := trace.NewFlightRecorder()
	if err := fr.Start(); err != nil {
		return status.Error(codes.Unavailable, err.Error())
	}
	defer errors.Invoke(&retErr, fr.Stop, "stop flightrecorder")

	time.Sleep(req.GetDuration().AsDuration())
	var buf bytes.Buffer
	if _, err := fr.WriteTo(&buf); err != nil {
		return status.Errorf(codes.Unknown, "snapshot flightrecorder: %v", err)
	}
	p := make([]byte, 10*units.MiB)
	for {
		n, err := buf.Read(p)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return status.Errorf(codes.Unknown, "read buffer: %v", err)
		}
		if err := s.Send(&debug.TraceChunk{
			Reply: &debug.TraceChunk_Bytes{
				Bytes: wrapperspb.Bytes(p[:n]),
			},
		}); err != nil {
			return status.Errorf(codes.Unknown, "send chunk: %v", err)
		}
	}
	return nil
}
