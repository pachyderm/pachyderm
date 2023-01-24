package main

import (
	"errors"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployApp(ctx *pulumi.Context) error {
	enterpriseKey := os.Getenv("PACH_ENTERPRISE_TOKEN")
	if enterpriseKey == "" {
		return errors.New("Need to supply env var PACH_ENTERPRISE_TOKEN")
	}
	return nil
}
