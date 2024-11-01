package dockertestenv

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	docker "github.com/docker/docker/client"
	minio "github.com/minio/minio-go/v7"
	minioCreds "github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

const (
	minioPort = 9002
)

// newTestMinioBucket creates a new bucket, which will be cleaned up when the test finishes
func newTestMinioBucket(ctx context.Context, t testing.TB, client *minio.Client) string {
	bucketName := testBucketName(t)
	t.Log("bucket:", bucketName)
	err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		err := client.RemoveBucketWithOptions(ctx, bucketName, minio.RemoveBucketOptions{ForceDelete: true})
		require.NoError(t, err)
	})
	return bucketName
}

func getMinioEndpoint() string {
	host := getDockerHost()
	return fmt.Sprintf("%s:%d", host, minioPort)
}

func testBucketName(t testing.TB) string {
	buf := make([]byte, 4)
	_, err := rand.Read(buf[:])
	require.NoError(t, err)
	tname := t.Name()
	tname = strings.ToLower(tname)
	tname = strings.ReplaceAll(tname, "/", "-")
	tname = strings.ReplaceAll(tname, "_", "-")
	return fmt.Sprintf("%s-%x", tname, buf[:])
}

var minioLock sync.Mutex

func ensureMinio(ctx context.Context, dclient docker.APIClient) error {
	const (
		containerName = "pach_test_minio_v2"
		imageName     = "minio/minio"
	)
	// bazel run //src/testing/cmd/dockertestenv creates these for many CI runs.
	if got, want := os.Getenv("SKIP_DOCKER_MINIO_CREATE"), "1"; got == want {
		log.Info(ctx, "not attempting to create docker container; SKIP_DOCKER_MINIO_CREATE=1")
		return nil
	}

	minioLock.Lock()
	defer minioLock.Unlock()
	return ensureContainer(ctx, dclient, containerName, containerSpec{
		Image: imageName,
		Cmd:   []string{"server", "/data", `--console-address=:9001`},
		PortMap: map[uint16]uint16{
			minioPort:     9000,
			minioPort + 1: 9001,
		},
	})
}

func NewMinioClient(ctx context.Context) (_ *minio.Client, retErr error) {
	dclient := newDockerClient()
	defer errors.Close(&retErr, dclient, "close docker client")
	if err := backoff.Retry(func() error {
		return ensureMinio(ctx, dclient)
	}, backoff.NewConstantBackOff(time.Second*3)); err != nil {
		return nil, errors.Wrap(err, "ensureMinio")
	}

	endpoint := getMinioEndpoint()
	id := "minioadmin"
	secret := "minioadmin"
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  minioCreds.NewStaticV4(id, secret, ""),
		Secure: false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "minio.New")
	}
	return client, nil
}

func NewTestBucket(ctx context.Context, t testing.TB) (*blob.Bucket, string) {
	t.Helper()
	client, err := NewMinioClient(ctx)
	require.NoError(t, err)
	bucketName := newTestMinioBucket(ctx, t, client)
	endpoint := getMinioEndpoint()
	id := "minioadmin"
	secret := "minioadmin"
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("dummy-region"),
		Credentials:      credentials.NewStaticCredentials(id, secret, ""),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	require.NoError(t, err)
	sctx, done := log.SpanContext(ctx, "OpenBucket", zap.String("bucket", bucketName))
	bucket, err := s3blob.OpenBucket(sctx, sess, bucketName, nil)
	done(zap.Error(err))
	require.NoError(t, err)
	return bucket, obj.ObjectStoreURL{
		Scheme: "minio",
		Bucket: fmt.Sprintf("%s/%s", endpoint, bucketName),
	}.String()
}

func NewTestBucketCtx(ctx context.Context) (*blob.Bucket, string, func(ctx context.Context) error, error) {
	client, err := NewMinioClient(ctx)
	if err != nil {
		return nil, "", nil, errors.Wrap(err, "get minio client")
	}
	endpoint := getMinioEndpoint()
	id := "minioadmin"
	secret := "minioadmin"
	buf := make([]byte, 4)
	if _, err := rand.Reader.Read(buf[:]); err != nil {
		return nil, "", nil, errors.Wrap(err, "generate bucket name: Read")
	}
	bucketName := fmt.Sprintf("%x", buf[:])
	if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
		return nil, "", nil, errors.Wrapf(err, "minio.MakeBucket(%v)", bucketName)
	}
	cleanupMinio := func(ctx context.Context) error {
		if err := client.RemoveBucketWithOptions(ctx, bucketName, minio.RemoveBucketOptions{ForceDelete: true}); err != nil {
			return errors.Wrapf(err, "RemoveBucket(%v)", bucketName)
		}
		return nil
	}

	sess, err := session.NewSession(&aws.Config{
		HTTPClient:       &http.Client{Transport: promutil.InstrumentRoundTripper("minio", nil)},
		Region:           aws.String("dummy-region"),
		Credentials:      credentials.NewStaticCredentials(id, secret, ""),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, "", cleanupMinio, errors.Wrap(err, "session.NewSession")
	}
	sctx, done := log.SpanContext(ctx, "OpenBucket", zap.String("bucket", bucketName))
	bucket, err := s3blob.OpenBucket(sctx, sess, bucketName, nil)
	done(zap.Error(err))
	if err != nil {
		return nil, "", cleanupMinio, errors.Wrap(err, "s3blob.OpenBucket")
	}
	return bucket, obj.ObjectStoreURL{
		Scheme: "minio",
		Bucket: fmt.Sprintf("%s/%s", endpoint, bucketName),
	}.String(), cleanupMinio, nil
}
