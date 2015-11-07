package server

import (
	"fmt"
	"io"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
	"google.golang.org/cloud/storage"

	"github.com/pachyderm/pachyderm/src/pfs"
)

type googleReplicaAPIServer struct {
	ctx    context.Context
	bucket string
}

func newGoogleReplicaAPIServer(ctx context.Context, bucket string) *googleReplicaAPIServer {
	return &googleReplicaAPIServer{ctx, bucket}
}

func (s *googleReplicaAPIServer) PullDiff(request *pfs.PullDiffRequest, pullDiffServer pfs.ReplicaAPI_PullDiffServer) (retErr error) {
	reader, err := storage.NewReader(s.ctx, s.bucket, fmt.Sprintf("%s/%s/%d", request.Commit.Repo.Name, request.Commit.Id, request.Shard))
	if err != nil {
		return err
	}
	defer func() {
		if err := reader.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	writer := protostream.NewStreamingBytesWriter(pullDiffServer)
	_, err = io.Copy(writer, reader)
	return err
}

func (s *googleReplicaAPIServer) PushDiff(pushDiffServer pfs.ReplicaAPI_PushDiffServer) (retErr error) {
	defer func() {
		if err := pushDiffServer.SendAndClose(google_protobuf.EmptyInstance); err != nil && retErr == nil {
			retErr = err
		}
	}()
	request, err := pushDiffServer.Recv()
	if err != nil {
		return err
	}
	writer := storage.NewWriter(s.ctx, s.bucket, fmt.Sprintf("%s/%s/%d", request.Commit.Repo.Name, request.Commit.Id, request.Shard))
	defer func() {
		if err := writer.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	reader := &pushDiffReader{
		server: pushDiffServer,
	}
	_, err = reader.buffer.Write(request.Value)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, reader)
	return err
}
