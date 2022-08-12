package integration

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	minio "github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, c *minio.Client, key, path string, size int64) {
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	_, err = c.PutObject("test-minio-go-lib", key, f, size, minio.PutObjectOptions{})
	require.NoError(t, err)
}

func compareFiles(t *testing.T, c *minio.Client, key, path string) {
	local, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	remoteObj, err := c.GetObject("test-minio-go-lib", key, minio.GetObjectOptions{})
	require.NoError(t, err)
	remote, err := ioutil.ReadAll(remoteObj)
	require.NoError(t, err)

	require.Equal(t, local, remote)
}

func listFiles(t *testing.T, c *minio.Client) []string {
	doneCh := make(chan struct{})
	defer close(doneCh)

	objectCh := c.ListObjectsV2("test-minio-go-lib", "", true, doneCh)
	files := []string{}
	for obj := range objectCh {
		require.NoError(t, obj.Err)
		if !strings.HasSuffix(obj.Key, "/") {
			files = append(files, obj.Key)
		}
	}

	return files
}

func TestMinioGoLib(t *testing.T) {
	netloc := os.Getenv("S2_HOST_NETLOC")
	accessKey := os.Getenv("S2_ACCESS_KEY")
	secretKey := os.Getenv("S2_SECRET_KEY")
	secure := os.Getenv("S2_HOST_SCHEME") == "https"
	c, err := minio.New(netloc, accessKey, secretKey, secure)
	require.NoError(t, err)

	require.NoError(t, c.MakeBucket("test-minio-go-lib", ""))

	writeFile(t, c, "small.txt", "../testdata/small.txt", 1)
	writeFile(t, c, "large.txt", "../testdata/large.txt", 65*1024*1024)

	require.ElementsMatch(t, listFiles(t, c), []string{"large.txt", "small.txt"})

	compareFiles(t, c, "small.txt", "../testdata/small.txt")
	compareFiles(t, c, "large.txt", "../testdata/large.txt")

	require.NoError(t, c.RemoveObject("test-minio-go-lib", "small.txt"))
	require.NoError(t, c.RemoveObject("test-minio-go-lib", "large.txt"))
	require.NoError(t, c.RemoveBucket("test-minio-go-lib"))
}
