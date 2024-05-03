package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/require"
)

type mocks struct{}

func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := args.Inputs.Mappable()
	switch args.TypeToken {
	case "aws:rds/instance:Instance":
		outputs["address"] = "postgres"
	}
	return args.Name + "_id", resource.NewPropertyMapFromMap(outputs), nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func TestDeployApp(t *testing.T) {
	chart, err := runfiles.Rlocation("_main/etc/helm/pachyderm")
	if err != nil {
		t.Logf("error resolving chart: %v, using etc/helm/pachyderm as the chart location", err)
		chart = "etc/helm/pachyderm"
	}
	if helmPath, ok := bazel.FindBinary("//tools/helm", "_helm"); ok {
		tmpdir := t.TempDir()
		if err := os.Symlink(helmPath, filepath.Join(tmpdir, "helm")); err != nil {
			t.Fatalf("symlink helm into $PATH: %v", err)
		}
		os.Setenv("PATH", fmt.Sprintf("%s:%s", tmpdir, os.Getenv("PATH")))
	}

	for _, stack := range []string{"dev1", "LOAD1", "LOAD2", "LOAD3", "LOAD4", "qa1", "qa2", "qa3", "qa4"} {
		t.Run(stack, func(t *testing.T) {
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				os.Setenv("BIGQUERY_AUTH_JSON", "{}")
				os.Setenv("CF_WP_LOADTEST_AWSKEYID", "x")
				os.Setenv("CF_WP_LOADTEST_AWSACCESSKEY", "x")
				os.Setenv("CF_WP_LOADTEST_ENDPOINT_URL", "s3://example.com")
				os.Setenv("ISSUER_URI", "https://example.com")
				os.Setenv("CLIENT_ID", "x")
				os.Setenv("CLIENT_SECRET", "x")
				os.Setenv("TLS_CRT", "x")
				os.Setenv("TLS_KEY", "x")

				rdsInstance, err := DeployRDS(ctx)
				if err != nil {
					return errors.Wrap(err, "XXX")
				}

				bucket, err := DeployBucket(ctx)
				if err != nil {
					return errors.Wrap(err, "DeployBucket")
				}

				renderProvider, err := kubernetes.NewProvider(ctx, "k8s-yaml-renderer", &kubernetes.ProviderArgs{
					RenderYamlToDirectory: pulumi.String(t.TempDir()),
				})
				if err != nil {
					return errors.Wrap(err, "NewProvider")
				}

				valuesOutput, err := DeployApp(ctx, renderProvider, nil, rdsInstance, bucket)
				if err != nil {
					return errors.Wrap(err, "DeployApp")
				}

				valuesCh := make(chan map[string]any)
				valuesOutput.ToMapOutput().ApplyT(func(x map[string]any) error {
					valuesCh <- x
					close(valuesCh)
					return nil
				})
				values := <-valuesCh
				js, err := json.Marshal(values)
				if err != nil {
					return errors.Wrap(err, "marshal helm values")
				}
				valuesFile := filepath.Join(t.TempDir(), "values.json")
				if err := os.WriteFile(valuesFile, js, 0o644); err != nil {
					return errors.Wrap(err, "write helm values")
				}

				helm.RenderTemplate(t, &helm.Options{
					ValuesFiles: []string{valuesFile},
					Logger:      logger.Discard,
				}, chart, "pachyderm", nil)

				return nil
			}, func(ri *pulumi.RunInfo) {
				pulumi.WithMocks("project", stack, mocks{})(ri)
				ri.Config = map[string]string{
					"project:rdsPGDBPassword": "x",
				}
			})
			require.NoError(t, err, "DeployApp should run")
		})
	}
}
