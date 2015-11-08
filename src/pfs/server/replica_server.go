package server

import (
	"bytes"
	"fmt"
	"io"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
	"google.golang.org/cloud/storage"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pachyderm/pachyderm/src/pfs"
)

var (
	maxChunkSize       = 1024 * 1024 * 1024 // 1 GB
	minChunkSize       = 5 * 1024 * 1024    // 5 MB
	multipartThreshold = 100 * 1024 * 1024  // 100 MB
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

type amazonReplicaAPIServer struct {
	s3     *s3.S3
	bucket string
}

func (s *amazonReplicaAPIServer) PullDiff(request *pfs.PullDiffRequest, pullDiffServer pfs.ReplicaAPI_PullDiffServer) (retErr error) {
	response, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    pullKey(request),
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := response.Body.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	writer := protostream.NewStreamingBytesWriter(pullDiffServer)
	_, err = io.Copy(writer, response.Body)
	return err
}

func (s *amazonReplicaAPIServer) PushDiff(pushDiffServer pfs.ReplicaAPI_PushDiffServer) (retErr error) {
	defer func() {
		if err := pushDiffServer.SendAndClose(google_protobuf.EmptyInstance); err != nil && retErr == nil {
			retErr = err
		}
	}()
	request, err := pushDiffServer.Recv()
	if err != nil {
		return err
	}
	writer := &s3Writer{
		s3:     s.s3,
		bucket: s.bucket,
		key:    pushKey(request),
	}
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

type s3Writer struct {
	s3         *s3.S3
	bucket     string
	key        *string
	buffer     bytes.Buffer
	uploadID   *string
	partNumber int64
	data       []byte // part of the class so we can lazily allocate
	parts      []*s3.CompletedPart
}

func (w *s3Writer) Write(data []byte) (int, error) {
	size, err := w.buffer.Write(data)
	if err != nil {
		return 0, err
	}
	if w.buffer.Len() < multipartThreshold {
		return size, nil
	}
	if w.uploadID == nil {
		response, err := w.s3.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
			Bucket: &w.bucket,
			Key:    w.key,
		})
		if err != nil {
			return 0, err
		}
		w.uploadID = response.UploadId
		w.data = make([]byte, maxChunkSize, maxChunkSize+minChunkSize)
	}
	// make sure we can send maxChunkSize and still have minChunkSize left in case they close it.
	if w.buffer.Len() > maxChunkSize+minChunkSize {
		if err := w.sendChunk(); err != nil {
			return 0, err
		}
	}
	return size, nil
}

func (w *s3Writer) Close() error {
	if w.buffer.Len() < multipartThreshold {
		w.data = make([]byte, minChunkSize)
		if _, err := w.buffer.Read(w.data); err != nil {
			return err
		}
		_, err := w.s3.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(w.data),
			Bucket: &w.bucket,
			Key:    w.key,
		})
		return err
	}
	w.data = w.data[:cap(w.data)]
	if err := w.sendChunk(); err != nil {
		return err
	}
	_, err := w.s3.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket: &w.bucket,
		Key:    w.key,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: w.parts,
		},
		UploadId: w.uploadID,
	})
	return err
}

func (w *s3Writer) sendChunk() error {
	if _, err := w.buffer.Read(w.data); err != nil {
		return err
	}
	partNumber := w.partNumber
	w.partNumber++
	response, err := w.s3.UploadPart(&s3.UploadPartInput{
		Body:       bytes.NewReader(w.data),
		Bucket:     &w.bucket,
		Key:        w.key,
		PartNumber: &partNumber,
	})
	if err != nil {
		return err
	}
	w.parts = append(w.parts, &s3.CompletedPart{
		ETag:       response.ETag,
		PartNumber: &partNumber,
	})
	return nil
}

func pullKey(request *pfs.PullDiffRequest) *string {
	result := fmt.Sprintf("%s/%s/%d", request.Commit.Repo.Name, request.Commit.Id, request.Shard)
	return &result
}

func pushKey(request *pfs.PushDiffRequest) *string {
	result := fmt.Sprintf("%s/%s/%d", request.Commit.Repo.Name, request.Commit.Id, request.Shard)
	return &result
}
