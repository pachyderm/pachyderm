package dockertestenv

import (
	"context"
	"crypto/rand"
	"fmt"
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
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

const (
	minioPort = 9000
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
	return fmt.Sprintf("%s-%x", tname, buf[:])
}

var minioLock sync.Mutex

func ensureMinio(ctx context.Context, dclient docker.APIClient) error {
	const (
		containerName = "pach_test_minio"
		imageName     = "minio/minio"
	)
	minioLock.Lock()
	defer minioLock.Unlock()
	return ensureContainer(ctx, dclient, containerName, containerSpec{
		Image: imageName,
		Cmd:   []string{"server", "/data", `--console-address=:9001`},
		PortMap: map[uint16]uint16{
			minioPort:     minioPort,
			minioPort + 1: minioPort + 1,
		},
	})
}

func NewTestBucket(ctx context.Context, t testing.TB) (*blob.Bucket, string) {
	t.Helper()
	dclient := newDockerClient()
	defer dclient.Close()
	err := backoff.Retry(func() error {
		return ensureMinio(ctx, dclient)
	}, backoff.NewConstantBackOff(time.Second*3))
	require.NoError(t, err)
	endpoint := getMinioEndpoint()
	id := "minioadmin"
	secret := "minioadmin"
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  minioCreds.NewStaticV4(id, secret, ""),
		Secure: false,
	})
	require.NoError(t, err)
	bucketName := newTestMinioBucket(ctx, t, client)
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("dummy-region"),
		Credentials:      credentials.NewStaticCredentials(id, secret, ""),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	require.NoError(t, err)
	bucket, err := s3blob.OpenBucket(ctx, sess, bucketName, nil)
	require.NoError(t, err)
	return bucket, obj.ObjectStoreURL{
		Scheme: "s3://",
		Bucket: fmt.Sprintf("s3://%s?endpoint=%s&disableSSL=true&region=dummy-region", endpoint, bucketName),
	}.String()
}
