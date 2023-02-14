package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(DeployResources())
}

func DeployResources() pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		k8sProvider, saRole, err := DeployCluster(ctx)
		if err != nil {
			return err
		}
		rdsInstance, err := DeployRDS(ctx)
		if err != nil {
			return err
		}
		bucket, err := DeployBucket(ctx)
		if err != nil {
			return err
		}
		err = DeployApp(ctx, k8sProvider, saRole, rdsInstance, bucket)
		if err != nil {
			return err
		}

		readmePath := fmt.Sprintf("./Pulumi.%s.README.md", ctx.Stack())
		if _, err := os.Stat(readmePath); err == nil {
			readmeBytes, err := ioutil.ReadFile(readmePath)
			if err != nil {
				return fmt.Errorf("failed to read readme: %w", err)
			}
			ctx.Export("readme", pulumi.String(string(readmeBytes)))
		} else {
			fmt.Printf("README file does not exist.\n")
		}

		return nil
	}
}
