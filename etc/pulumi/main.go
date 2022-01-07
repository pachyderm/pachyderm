package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/storage"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	helm "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		slug := "pachyderm/ci-cluster/dev"
		stackRef, _ := pulumi.NewStackReference(ctx, slug, nil)

		kubeConfig := stackRef.GetOutput(pulumi.String("kubeconfig"))
		ipAddress := stackRef.GetOutput(pulumi.String("ingressIP"))
		sha := cfg.Require("sha")
		//url := cfg.Require("url")

		k8sProvider, err := kubernetes.NewProvider(ctx, "k8sprovider", &kubernetes.ProviderArgs{
			Kubeconfig: pulumi.StringOutput(kubeConfig),
		}) //, pulumi.DependsOn([]pulumi.Resource{cluster})
		if err != nil {
			return err
		}

		bucket, err := storage.NewBucket(ctx, "pach-bucket", &storage.BucketArgs{
			Name:     pulumi.String(sha),
			Location: pulumi.String("US"),
			/*
				Labels: pulumi.StringMap{
					"workspace": pulumi.String("myfinecluster"),
				},
			*/
		})
		if err != nil {
			return err
		}

		//TODO Create Service account for each pach install and assign to bucket
		defaultSA := compute.GetDefaultServiceAccountOutput(ctx, compute.GetDefaultServiceAccountOutputArgs{}, nil)
		if err != nil {
			return err
		}

		_, err = storage.NewBucketIAMMember(ctx, "bucket-role", &storage.BucketIAMMemberArgs{
			Bucket: bucket.Name,
			Role:   pulumi.String("roles/storage.admin"),
			Member: defaultSA.Email().ApplyT(func(s string) string { return "serviceAccount:" + s }).(pulumi.StringOutput),
		})
		if err != nil {
			return err
		}

		namespace, err := corev1.NewNamespace(ctx, "pach-ns", &corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(sha),
				Labels: pulumi.StringMap{
					"needs-ci-tls": pulumi.String("true"), //Uses kubernetes replicator to replicate TLS secret to new NS
				},
			},
		}, pulumi.Provider(k8sProvider))

		if err != nil {
			return err
		}

		_, err = helm.NewRelease(ctx, "pach-release", &helm.ReleaseArgs{
			Namespace: namespace.Metadata.Elem().Name(),
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://helm.pachyderm.com"), //TODO Use Chart files in Repo
			},
			Chart: pulumi.String("pachyderm"),
			Values: pulumi.Map{
				"deployTarget": pulumi.String("GOOGLE"),
				"console": pulumi.Map{
					"enabled": pulumi.Bool(true),
				},
				"ingress": pulumi.Map{
					"annotations": pulumi.Map{
						"kubernetes.io/ingress.class":              pulumi.String("traefik"),
						"traefik.ingress.kubernetes.io/router.tls": pulumi.String("true"),
					},
					"enabled": pulumi.Bool(true),
					"host":    pulumi.String(sha + ".clusters-ci.pachyderm.io"),
					"tls": pulumi.Map{
						"enabled":    pulumi.Bool(true),
						"secretName": pulumi.String("wildcard-tls"), // Dynamic Value
					},
				},
				"pachd": pulumi.Map{
					"externalService": pulumi.Map{
						"enabled":        pulumi.Bool(true),
						"loadBalancerIP": ipAddress,
						"apiGRPCPort":    pulumi.Int(30650), //Dynamic Value
						"s3GatewayPort":  pulumi.Int(30600), //Dynamic Value
					},
					"enterpriseLicenseKey": cfg.RequireSecret("entActCode"), //Set in .circleci/config.yml
					"storage": pulumi.Map{
						"google": pulumi.Map{
							"bucket": bucket.Name,
						},
						"tls": pulumi.Map{
							"enabled":    pulumi.Bool(true),
							"secretName": pulumi.String("wildcard-tls"),
						},
					},
				},
			},
		}, pulumi.Provider(k8sProvider))

		if err != nil {
			return err
		}

		return nil
	})
}
