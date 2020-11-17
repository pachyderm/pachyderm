package testing

import (
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

func NewDefaultAmazonConfig() *obj.AmazonAdvancedConfiguration {
	return &obj.AmazonAdvancedConfiguration{
		Retries:        obj.DefaultRetries,
		Timeout:        obj.DefaultTimeout,
		UploadACL:      obj.DefaultUploadACL,
		Reverse:        obj.DefaultReverse,
		PartSize:       obj.DefaultPartSize,
		MaxUploadParts: obj.DefaultMaxUploadParts,
		DisableSSL:     obj.DefaultDisableSSL,
		NoVerifySSL:    obj.DefaultNoVerifySSL,
		LogOptions:     obj.DefaultAwsLogOptions,
	}
}

func LoadAmazonParameters(t *testing.T) (string, string, string, string) {
	id := os.Getenv("AMAZON_DEPLOYMENT_ID")
	secret := os.Getenv("AMAZON_DEPLOYMENT_SECRET")
	bucket := os.Getenv("AMAZON_DEPLOYMENT_BUCKET")
	region := os.Getenv("AMAZON_DEPLOYMENT_REGION")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", region)

	return id, secret, bucket, region
}

func LoadECSParameters(t *testing.T) (string, string, string, string, string) {
	id := os.Getenv("ECS_DEPLOYMENT_ID")
	secret := os.Getenv("ECS_DEPLOYMENT_SECRET")
	bucket := os.Getenv("ECS_DEPLOYMENT_BUCKET")
	endpoint := os.Getenv("ECS_DEPLOYMENT_CUSTOM_ENDPOINT")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", endpoint)

	return id, secret, bucket, "dummy-region", endpoint
}

func LoadGoogleParameters(t *testing.T) (string, string) {
	bucket := os.Getenv("GOOGLE_DEPLOYMENT_BUCKET")
	creds := os.Getenv("GOOGLE_DEPLOYMENT_CREDS")
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", creds)

	return bucket, creds
}

func LoadGoogleHMACParameters(t *testing.T) (string, string, string, string, string) {
	id := os.Getenv("GOOGLE_DEPLOYMENT_HMAC_ID")
	secret := os.Getenv("GOOGLE_DEPLOYMENT_HMAC_SECRET")
	bucket := os.Getenv("GOOGLE_DEPLOYMENT_BUCKET")
	region := os.Getenv("GOOGLE_DEPLOYMENT_REGION")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", region)

	return id, secret, bucket, region, "storage.googleapis.com"
}

func LoadMicrosoftParameters(t *testing.T) (string, string, string) {
	id := os.Getenv("MICROSOFT_DEPLOYMENT_ID")
	secret := os.Getenv("MICROSOFT_DEPLOYMENT_SECRET")
	container := os.Getenv("MICROSOFT_DEPLOYMENT_CONTAINER")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", container)

	return id, secret, container
}
