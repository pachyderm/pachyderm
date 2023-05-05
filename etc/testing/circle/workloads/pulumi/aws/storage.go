package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployBucket(ctx *pulumi.Context) (*s3.Bucket, error) {
	bucketName := strings.ToLower(fmt.Sprintf("s3-%s-bucket", ctx.Stack()))
	bucket, err := s3.NewBucket(ctx, bucketName, &s3.BucketArgs{
		Bucket:       pulumi.String(bucketName),
		ForceDestroy: pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Project":     pulumi.String("Feature Testing"),
			"Service":     pulumi.String("CI"),
			"Owner":       pulumi.String("pachyderm-ci"),
			"Team":        pulumi.String("Core"),
			"Environment": pulumi.String(ctx.Stack()),
		},
	})

	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("error creating S3 bucket: %w", err))
	}

	ctx.Export("bucketName", bucket.Bucket)

	return bucket, nil
}
