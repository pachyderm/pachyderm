package server

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"gocloud.dev/blob"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj/integrationtests"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestMicrosoft(t *testing.T) {
	integrationtests.LoadMicrosoftParameters(t)
	require.NoError(t, os.Setenv("AZURE_STORAGE_ACCOUNT", os.Getenv("MICROSOFT_CLIENT_ID")))
	require.NoError(t, os.Setenv("AZURE_STORAGE_KEY", os.Getenv("MICROSOFT_CLIENT_SECRET")))
	bucketName := os.Getenv("MICROSOFT_CLIENT_CONTAINER")
	url, err := obj.ParseURL("azblob://" + bucketName)
	require.NoError(t, err, "should be able to parse url")
	writeReadDelete(t, url)
	url, err = obj.ParseURL("wasb://" + bucketName + "@" + os.Getenv("MICROSOFT_CLIENT_ID") + ".blob.core.windows.net")
	require.NoError(t, err, "should be able to parse url")
	writeReadDelete(t, url)
}

func TestGoogle(t *testing.T) {
	integrationtests.LoadGoogleParameters(t)
	credFile := path.Join(t.TempDir(), "tmp-google-cred")
	require.NoError(t, os.WriteFile(credFile, []byte(os.Getenv("GOOGLE_CLIENT_CREDS")), 0666))
	require.NoError(t, os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credFile))
	bucketName := os.Getenv("GOOGLE_CLIENT_BUCKET")
	url, err := obj.ParseURL("gs://" + bucketName)
	require.NoError(t, err, "should be able to parse url")
	writeReadDelete(t, url)
}

func TestAmazon(t *testing.T) {
	integrationtests.LoadAmazonParameters(t)
	require.NoError(t, os.Setenv("AWS_REGION", os.Getenv("AMAZON_CLIENT_REGION")))
	require.NoError(t, os.Setenv("AWS_SECRET_ACCESS_KEY", os.Getenv("AMAZON_CLIENT_SECRET")))
	require.NoError(t, os.Setenv("AWS_ACCESS_KEY_ID", os.Getenv("AMAZON_CLIENT_ID")))
	bucketName := os.Getenv("AMAZON_CLIENT_BUCKET")
	url, err := obj.ParseURL("s3://" + bucketName)
	require.NoError(t, err, "should be able to parse url")
	writeReadDelete(t, url)
}

func TestGoogleHMAC(t *testing.T) {
	integrationtests.LoadGoogleHMACParameters(t)
	require.NoError(t, os.Setenv("AWS_REGION", os.Getenv("GOOGLE_CLIENT_REGION")))
	require.NoError(t, os.Setenv("AWS_SECRET_ACCESS_KEY", os.Getenv("GOOGLE_CLIENT_HMAC_SECRET")))
	require.NoError(t, os.Setenv("AWS_ACCESS_KEY_ID", os.Getenv("GOOGLE_CLIENT_HMAC_ID")))
	require.NoError(t, os.Setenv("CUSTOM_ENDPOINT", "storage.googleapis.com"))
	bucketName := os.Getenv("GOOGLE_CLIENT_BUCKET")
	url, err := obj.ParseURL("s3://" + bucketName)
	require.NoError(t, err, "should be able to parse url")
	writeReadDelete(t, url)
}

func TestAmazonECS(t *testing.T) {
	t.Skip("Skip until ECS is available and stable.")
	integrationtests.LoadECSParameters(t)
	require.NoError(t, os.Setenv("AWS_REGION", "dummy-region"))
	require.NoError(t, os.Setenv("AWS_SECRET_ACCESS_KEY", os.Getenv("ECS_CLIENT_SECRET")))
	require.NoError(t, os.Setenv("AWS_ACCESS_KEY_ID", os.Getenv("ECS_CLIENT_ID")))
	require.NoError(t, os.Setenv("CUSTOM_ENDPOINT", os.Getenv("ECS_CLIENT_CUSTOM_ENDPOINT")))
	// We may have to check if we have valid ssl for our ECS configuration.
	require.NoError(t, os.Setenv("DISABLE_SSL", "true"))
	bucketName := os.Getenv("ECS_CLIENT_BUCKET")
	url, err := obj.ParseURL("s3://" + bucketName)
	require.NoError(t, err, "should be able to parse url")
	writeReadDelete(t, url)
}

func writeReadDelete(t *testing.T, url *obj.ObjectStoreURL) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	bucket, err := openBucket(ctx, url)
	defer func() {
		require.NoError(t, bucket.Close())
	}()
	require.NoError(t, err, "should be able to open bucket")
	objName := randutil.UniqueString("test-object-")
	writeToObjStorage(ctx, t, bucket, objName)
	buf := &bytes.Buffer{}
	readFromObjStorage(ctx, t, bucket, objName, buf)
	require.Equal(t, buf.String(), objName)
	require.NoError(t, bucket.Delete(ctx, objName), "should be able to delete object")
}

func writeToObjStorage(ctx context.Context, t *testing.T, bucket *blob.Bucket, objName string, data ...string) {
	exists, err := bucket.Exists(ctx, objName)
	require.NoError(t, err, fmt.Sprintf("should be able to check if obj %s exists", objName))
	require.Equal(t, false, exists)
	w, err := bucket.NewWriter(ctx, objName, nil)
	require.NoError(t, err, fmt.Sprintf("should be able to create writer for %s", objName))
	defer func() {
		if err := w.Close(); err != nil {
			require.NoError(t, err, "should be able to close writer")
		}
	}()
	objData := []byte(objName)
	if len(data) != 0 {
		objData = []byte(data[0])
	}
	_, err = w.Write(objData)
	require.NoError(t, err, fmt.Sprintf("should be able to write to %s", objName))
}

func readFromObjStorage(ctx context.Context, t *testing.T, bucket *blob.Bucket, objName string, buf *bytes.Buffer) {
	r, err := bucket.NewReader(ctx, objName, nil)
	require.NoError(t, err, fmt.Sprintf("should be able to create reader for obj %s", objName))
	defer func() {
		if err := r.Close(); err != nil {
			require.NoError(t, err, "should be able to close reader")
		}
	}()
	_, err = r.WriteTo(buf)
	require.NoError(t, err, fmt.Sprintf("should be able to read from obj %s", objName))
}
