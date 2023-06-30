package obj

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
)

func NewAmazonBucketFromEnv(ctx context.Context) (*Bucket, error) {
	region, ok := os.LookupEnv(AmazonRegionEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", AmazonRegionEnvVar)
	}
	bucket, ok := os.LookupEnv(AmazonBucketEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", AmazonBucketEnvVar)
	}

	var creds AmazonCreds
	creds.ID, _ = os.LookupEnv(AmazonIDEnvVar)
	creds.Secret, _ = os.LookupEnv(AmazonSecretEnvVar)
	creds.Token, _ = os.LookupEnv(AmazonTokenEnvVar)

	// Get endpoint for custom deployment (optional).
	endpoint, _ := os.LookupEnv(CustomEndpointEnvVar)
	return NewAmazonBucket(ctx, s3.Options{
		Region:           region,
		EndpointResolver: s3.EndpointResolverFromURL(endpoint),
	}, bucket)
}

// NewAmazonBucketFromSecret constructs an amazon client by reading credentials
// from a mounted AmazonSecret. You may pass "" for bucket in which case it
// will read the bucket from the secret.
func NewAmazonBucketFromSecret(ctx context.Context, bucket string) (*Bucket, error) {
	// Get AWS region (required for constructing an AWS client)
	region, err := readSecretFile(fmt.Sprintf("/%s", AmazonRegionEnvVar))
	if err != nil {
		return nil, errors.Errorf("amazon-region not found")
	}

	// Use or retrieve S3 bucket
	if bucket == "" {
		bucket, err = readSecretFile(fmt.Sprintf("/%s", AmazonBucketEnvVar))
		if err != nil {
			return nil, err
		}
	}

	// Retrieve static credentials; if not found,
	// use IAM roles (i.e. the EC2 metadata service)
	var creds AmazonCreds
	creds.ID, err = readSecretFile(fmt.Sprintf("/%s", AmazonIDEnvVar))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	creds.Secret, err = readSecretFile(fmt.Sprintf("/%s", AmazonSecretEnvVar))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	creds.Token, err = readSecretFile(fmt.Sprintf("/%s", AmazonTokenEnvVar))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// Get endpoint for custom deployment (optional).
	endpoint, err := readSecretFile(fmt.Sprintf("/%s", CustomEndpointEnvVar))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return NewAmazonBucket(ctx, s3.Options{
		Region:           region,
		EndpointResolver: s3.EndpointResolverFromURL(endpoint),
	}, bucket)
}

// NewGoogleBucketFromSecret creates a google Bucket by reading credentials
// from a mounted GoogleSecret. You may pass "" for bucket in which case it
// will read the bucket from the secret.
func NewGoogleBucketFromSecret(ctx context.Context, bucket string) (*Bucket, error) {
	var err error
	if bucket == "" {
		bucket, err = readSecretFile(fmt.Sprintf("/%s", GoogleBucketEnvVar))
		if err != nil {
			return nil, errors.Errorf("google-bucket not found")
		}
	}
	credData, err := readSecretFile(fmt.Sprintf("/%s", GoogleCredEnvVar))
	if err != nil {
		return nil, errors.Errorf("google-cred not found")
	}
	var creds *google.Credentials
	if credData != "" {
		jsonData, err := os.ReadFile(secretFile(fmt.Sprintf("/%s", GoogleCredEnvVar)))
		if err != nil {
			return nil, err
		}
		creds, err = google.CredentialsFromJSON(ctx, jsonData)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		creds, err = google.FindDefaultCredentials(ctx)
		if err != nil {
			return nil, err
		}
	}
	return NewGoogleBucket(ctx, creds, bucket)
}

func NewGoogleBucketFromEnv(ctx context.Context) (*Bucket, error) {
	bucket, ok := os.LookupEnv(GoogleBucketEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", GoogleBucketEnvVar)
	}
	credData, ok := os.LookupEnv(GoogleCredEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", GoogleCredEnvVar)
	}
	creds, err := google.CredentialsFromJSON(ctx, []byte(credData))
	if err != nil {
		return nil, err
	}
	return NewGoogleBucket(ctx, creds, bucket)
}

// NewMicrosoftBucketFromSecret creates a microsoft client by reading
// credentials from a mounted MicrosoftSecret. You may pass "" for container in
// which case it will read the container from the secret.
func NewMicrosoftBucketFromSecret(ctx context.Context, container string) (*Bucket, error) {
	var err error
	if container == "" {
		container, err = readSecretFile(fmt.Sprintf("/%s", MicrosoftContainerEnvVar))
		if err != nil {
			return nil, errors.Errorf("microsoft-container not found")
		}
	}
	id, err := readSecretFile(fmt.Sprintf("/%s", MicrosoftIDEnvVar))
	if err != nil {
		return nil, errors.Errorf("microsoft-id not found")
	}
	secret, err := readSecretFile(fmt.Sprintf("/%s", MicrosoftSecretEnvVar))
	if err != nil {
		return nil, errors.Errorf("microsoft-secret not found")
	}
	// TODO
	log.Debug(ctx, "", zap.String("id", id), zap.String("secret", secret))
	return NewMicrosoftBucket(ctx, "", nil, container)
}

// NewMicrosoftBucketFromEnv creates a Microsoft client based on environment variables.
func NewMicrosoftBucketFromEnv(ctx context.Context) (*Bucket, error) {
	container, ok := os.LookupEnv(MicrosoftContainerEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MicrosoftContainerEnvVar)
	}
	id, ok := os.LookupEnv(MicrosoftIDEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MicrosoftIDEnvVar)
	}
	secret, ok := os.LookupEnv(MicrosoftSecretEnvVar)
	if !ok {
		return nil, errors.Errorf("%s not found", MicrosoftSecretEnvVar)
	}
	log.Debug(ctx, "", zap.String("id", id), zap.String("secret", secret))
	return NewMicrosoftBucket(ctx, "", nil, container)
}
