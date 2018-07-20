package server

import (
	"fmt"
	"runtime/pprof"

	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

func NewDebugServer() debug.DebugServer {
	return &DebugServer{}
}

type DebugServer struct {
}

func (s *DebugServer) Goro(request *debug.GoroRequest, server debug.Debug_GoroServer) error {
	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return fmt.Errorf("unable to find goroutine profile")
	}
	return profile.WriteTo(grpcutil.NewStreamingBytesWriter(server), 2)
}
