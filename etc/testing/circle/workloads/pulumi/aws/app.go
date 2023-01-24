package main

import (
	"errors"
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/rds"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func DeployApp(ctx *pulumi.Context, cluster *eks.Cluster, rdsInstance *rds.Instance, bucket *s3.Bucket) error {
	cfg := config.New(ctx, "")
	enterpriseKey := os.Getenv("ENT_ACT_TOKEN")
	if enterpriseKey == "" {
		return errors.New("Need to supply env var ENT_ACT_TOKEN")
	}
	awsSAkey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSAsecret := os.Getenv("AWS_SECRET_ACCESS_KEY")

	k8sProvider, err := kubernetes.NewProvider(ctx, "k8sprovider", &kubernetes.ProviderArgs{
		Kubeconfig: cluster.Kubeconfig.AsStringOutput(),
	})
	if err != nil {
		return err
	}

	namespace, err := corev1.NewNamespace(ctx, "test-ns", &corev1.NamespaceArgs{},
		pulumi.Provider(k8sProvider))

	if err != nil {
		return err
	}

	if enterpriseKey == "" {
		return errors.New("Need to supply env var PACH_ENTERPRISE_TOKEN")
	}

	_, err = helm.NewRelease(ctx, "pach-release", &helm.ReleaseArgs{
		Namespace: namespace.Metadata.Elem().Name(),
		RepositoryOpts: helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://helm.pachyderm.com"), //TODO Use Chart files in Core Pach Repo instead of latest released version
		},
		Chart: pulumi.String("pachyderm"),
		Values: pulumi.Map{
			"ingress": pulumi.Map{
				"enabled": pulumi.Bool(true),
				"host":    pulumi.String(""),
				// "annotations": pulumi.Map{
				// 	"kubernetes.io/ingress.class": pulumi.String("traefik"),
				// 	//"traefik.ingress.kubernetes.io/router.tls": "true",
				// },
			},
			"proxy": pulumi.Map{
				"enabled":  pulumi.Bool(true),
				"host":     pulumi.String(""),
				"replicas": pulumi.Int(1),
				"image": pulumi.Map{
					"repository": pulumi.String("envoyproxy/envoy-distroless"),
					"tag":        pulumi.String("v1.24.1"),
					"pullPolicy": pulumi.String("IfNotPresent"),
				},
			},
			"console": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"pachd": pulumi.Map{
				"storage": pulumi.Map{
					"amazon": pulumi.Map{
						"bucket": bucket.Bucket,
						"region": pulumi.String("us-west-2"),
						"id":     pulumi.String(awsSAkey),
						"secret": pulumi.String(awsSAsecret),
					},
				},
				"externalService": pulumi.Map{
					"enabled": pulumi.Bool(true),
				},
				"enterpriseLicenseKey": pulumi.String(enterpriseKey),
				"oauthClientSecret":    pulumi.String("i9mRbLujCvi8j3NPKOFPklXai71oqz3y"),
				"rootToken":            pulumi.String("1WgTXSc2MccsxunEzXvSAejsKNyT4Lsy"),
				"enterpriseSecret":     pulumi.String("SBgvzhmVtMxiVbzSIzpWqi3fKCfsup3o"),
			},
			"deployTarget": pulumi.String("AMAZON"),
			"global": pulumi.Map{
				"postgresql": pulumi.Map{
					"postgresqlHost":                   rdsInstance.Address,
					"postgresqlUsername":               pulumi.String("postgres"),
					"postgresqlPassword":               cfg.RequireSecret("rdsPGDBPassword"), //To allow for clean upgrade
					"postgresqlPostgresPassword":       cfg.RequireSecret("rdsPGDBPassword"),
					"identityDatabaseFullNameOverride": pulumi.String("dex"),
				},
			},
			"postgresql": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
		},
	}, pulumi.Provider(k8sProvider))

	if err != nil {
		return err
	}
	return nil

}
