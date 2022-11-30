package main

import (
	"io"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gogo/protobuf/types"
	"github.com/jrockway/opinionated-server/server"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type S struct {
	pfs.UnimplementedAPIServer
}

func (*S) ModifyFile(srv pfs.API_ModifyFileServer) error {
	files := map[string]int{}
	for {
		msg, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				for name, size := range files {
					zap.L().Info("added file", zap.String("name", name), zap.String("size", humanize.Bytes(uint64(size))))
				}
				return srv.SendAndClose(&types.Empty{})
			}
			return err
		}
		switch msg.GetBody().(type) {
		case *pfs.ModifyFileRequest_AddFile:
			path := msg.GetAddFile().Path
			add := len(msg.GetAddFile().GetRaw().GetValue())
			files[path] += add
			zap.L().Info("add file", zap.String("path", path), zap.String("size", humanize.Bytes(uint64(add))))
			time.Sleep(10 * time.Millisecond)
		default:
			zap.L().Info("other", zap.Any("data", msg))
		}
	}
}

func main() {
	server.AppName = "fake-pfs"
	server.Setup()

	pfsServer := new(S)

	server.AddService(func(s *grpc.Server) {
		pfs.RegisterAPIServer(s, pfsServer)
	})
	server.ListenAndServe()
}
