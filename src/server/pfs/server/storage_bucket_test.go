//go:build unit_test

package server

import (
	"bytes"
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

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
	readWriteDelete(t, url, bucketName)
}

func TestGoogle(t *testing.T) {
	integrationtests.LoadGoogleParameters(t)
	credFile := path.Join(t.TempDir(), "tmp-google-cred")
	require.NoError(t, os.WriteFile(credFile, []byte(os.Getenv("GOOGLE_CLIENT_CREDS")), 0666))
	require.NoError(t, os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credFile))
	bucketName := os.Getenv("GOOGLE_CLIENT_BUCKET")
	url, err := obj.ParseURL("gs://" + bucketName)
	require.NoError(t, err, "should be able to parse url")
	readWriteDelete(t, url, bucketName)
}

func TestAmazon(t *testing.T) {
	integrationtests.LoadAmazonParameters(t)
	require.NoError(t, os.Setenv("AWS_REGION", os.Getenv("AMAZON_CLIENT_REGION")))
	require.NoError(t, os.Setenv("AWS_SECRET_ACCESS_KEY", os.Getenv("AMAZON_CLIENT_SECRET")))
	require.NoError(t, os.Setenv("AWS_ACCESS_KEY_ID", os.Getenv("AMAZON_CLIENT_ID")))
	bucketName := os.Getenv("AMAZON_CLIENT_BUCKET")
	url, err := obj.ParseURL("s3://" + bucketName)
	require.NoError(t, err, "should be able to parse url")
	readWriteDelete(t, url, bucketName)
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
	readWriteDelete(t, url, bucketName)
}

func TestAmazonECS(t *testing.T) {
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
	readWriteDelete(t, url, bucketName)
}

func readWriteDelete(t *testing.T, url *obj.ObjectStoreURL, bucketName string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30000)
	defer cancel()
	bucket, err := openBucket(ctx, url)
	defer func() {
		require.NoError(t, bucket.Close())
	}()
	require.NoError(t, err, "should be able to open bucket")
	objName := randutil.UniqueString("test-object-")
	r := strings.NewReader(objName)
	err = exportObj(ctx, r, objName, bucketName, bucket)
	require.NoError(t, err, "should be able to export object")
	buf := bytes.NewBuffer(make([]byte, 0))
	err = importObj(ctx, buf, objName, bucketName, bucket)
	require.NoError(t, err, "should be able to import object")
	require.Equal(t, buf.String(), objName)
	err = bucket.Delete(ctx, objName)
	require.NoError(t, err, "should be able to delete object")
}
