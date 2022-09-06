package dockertestenv

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	docker "github.com/docker/docker/client"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

const (
	minioPort = 9000
)

func NewTestObjClient(t testing.TB) obj.Client {
	dclient := newDockerClient()
	defer dclient.Close()
	err := backoff.Retry(func() error {
		return ensureMinio(context.Background(), dclient)
	}, backoff.NewConstantBackOff(time.Second*3))
	require.NoError(t, err)
	endpoint := getMinioEndpoint()
	id := "minioadmin"
	secret := "minioadmin"
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(id, secret, ""),
		Secure: false,
	})
	require.NoError(t, err)
	bucketName := newTestMinioBucket(t, client)
	oc, err := obj.NewMinioClient(endpoint, bucketName, id, secret, false, false)
	require.NoError(t, err)
	return obj.WrapWithTestURL(oc)
}

// newTestMinioBucket creates a new bucket, which will be cleaned up when the test finishes
func newTestMinioBucket(t testing.TB, client *minio.Client) string {
	ctx := context.Background()
	bucketName := testBucketName(t)
	t.Log("bucket:", bucketName)
	err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		err := client.RemoveBucketWithOptions(ctx, bucketName, minio.BucketOptions{ForceDelete: true})
		require.NoError(t, err)
	})
	return bucketName
}

func getMinioEndpoint() string {
	host := getDockerHost()
	return fmt.Sprintf("%s:%d", host, minioPort)
}

func testBucketName(t testing.TB) string {
	buf := make([]byte, 8)
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
