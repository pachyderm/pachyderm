package testing

import (
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// The Load.*Parameters functions in this file are where we get the credentials
// for running object client tests. These are sourced from your environment
// variables. You can provide these yourself, or Pachyderm employees may use the
// variables defined in the password manager under the secure note "Pachyderm CI
// Object Storage Credentials". These should all be scoped to the smallest set
// of permissions necessary, which is just object create/read/delete within a
// specific bucket.

// LoadAmazonParameters loads the test parameters for S3 object storage:
//  id - the key id credential
//  secret - the key secret credential
//  bucket - the S3 bucket to issue requests towards
//  region - the S3 region that the bucket is in
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

// LoadECSParameters loads the test parameters for ECS object storage
// (credentials can be requested for a new account from portal.ecstestdrive.com):
//  id - the key id credential
//  secret - the key secret credential
//  bucket - the ECS bucket to issue requests towards
//  region - a dummy region - some clients require this but it is unused with 'endpoint'
//  endpoint - the S3-compatible server to send requests to
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

// LoadGoogleParameters loads the test parameters for GCS object storage:
//  bucket - the GCS bucket to issue requests towards
//  creds - the JSON GCP credentials to use
func LoadGoogleParameters(t *testing.T) (string, string) {
	bucket := os.Getenv("GOOGLE_DEPLOYMENT_BUCKET")
	creds := os.Getenv("GOOGLE_DEPLOYMENT_CREDS")
	require.NotEqual(t, "", bucket)
	require.NotEqual(t, "", creds)

	return bucket, creds
}

// LoadGoogleHMACParameters loads the test parameters for GCS object storage,
// specifically for use with S3-compatible clients:
//  id - the key id credential
//  secret - the key secret credential
//  bucket - the GCS bucket to issue requests towards
//  region - the GCS region that the bucket is in
//  endpoint - the S3-compatible server to send requests to
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

// LoadMicrosoftParameters loads the test parameters for Azure blob storage:
//  id - the key id credential
//  secret - the key secret credential
//  container - the Azure blob container to issue requests towards
func LoadMicrosoftParameters(t *testing.T) (string, string, string) {
	id := os.Getenv("MICROSOFT_DEPLOYMENT_ID")
	secret := os.Getenv("MICROSOFT_DEPLOYMENT_SECRET")
	container := os.Getenv("MICROSOFT_DEPLOYMENT_CONTAINER")
	require.NotEqual(t, "", id)
	require.NotEqual(t, "", secret)
	require.NotEqual(t, "", container)

	return id, secret, container
}
