package main

import (
	"fmt"
	"io/ioutil"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(DeployResources())
}

func DeployResources() pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		err := DeployCluster(ctx)
		if err != nil {
			return err
		}
		err = DeployRDS(ctx)
		if err != nil {
			return err
		}
		err = DeployBucket(ctx)
		if err != nil {
			return err
		}
		err = DeployApp(ctx)
		if err != nil {
			return err
		}

		readmePath := fmt.Sprintf("./Pulumi.%s.README.md", ctx.Stack())
		readmeBytes, err := ioutil.ReadFile(readmePath)
		if err != nil {
			return fmt.Errorf("failed to read readme: %w", err)
		}

		ctx.Export("readme", pulumi.String(string(readmeBytes)))
		return nil
	}
}
