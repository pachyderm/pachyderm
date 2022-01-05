package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/storage"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	helm "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		containerService, err := projects.NewService(ctx, "project", &projects.ServiceArgs{
			Service: pulumi.String("container.googleapis.com"),
		})
		if err != nil {
			return err
		}

		engineVersions, err := container.GetEngineVersions(ctx, &container.GetEngineVersionsArgs{})
		if err != nil {
			return err
		}
		masterVersion := engineVersions.LatestMasterVersion

		cluster, err := container.NewCluster(ctx, "ci-cluster", &container.ClusterArgs{
			InitialNodeCount: pulumi.Int(2),
			MinMasterVersion: pulumi.String(masterVersion),
			NodeVersion:      pulumi.String(masterVersion),
			NodeConfig: &container.ClusterNodeConfigArgs{
				MachineType: pulumi.String("n1-standard-1"),
				OauthScopes: pulumi.StringArray{
					pulumi.String("https://www.googleapis.com/auth/compute"),
					pulumi.String("https://www.googleapis.com/auth/devstorage.read_write"), //TODO Change back to read-only
					pulumi.String("https://www.googleapis.com/auth/logging.write"),
					pulumi.String("https://www.googleapis.com/auth/monitoring"),
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{containerService}))
		if err != nil {
			return err
		}

		ctx.Export("kubeconfig", generateKubeconfig(cluster.Endpoint, cluster.Name, cluster.MasterAuth))

		k8sProvider, err := kubernetes.NewProvider(ctx, "k8sprovider", &kubernetes.ProviderArgs{
			Kubeconfig: generateKubeconfig(cluster.Endpoint, cluster.Name, cluster.MasterAuth),
		}, pulumi.DependsOn([]pulumi.Resource{cluster}))
		if err != nil {
			return err
		}

		//TODO Static IP for Load Balancer
		_, err = helm.NewRelease(ctx, "traefik", &helm.ReleaseArgs{
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://helm.traefik.io/traefik"),
			},
			Chart: pulumi.String("traefik"),
		}, pulumi.Provider(k8sProvider))

		if err != nil {
			return err
		}

		_, err = helm.NewRelease(ctx, "cert-manager", &helm.ReleaseArgs{
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://charts.jetstack.io"),
			},
			Chart: pulumi.String("cert-manager"),
			Values: pulumi.Map{
				"installCRDs": pulumi.String("true"),
			},
		}, pulumi.Provider(k8sProvider))

		if err != nil {
			return err
		}

		_, err = helm.NewRelease(ctx, "kube-replicator", &helm.ReleaseArgs{
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://helm.mittwald.de"),
			},
			Chart: pulumi.String("kubernetes-replicator"),
		}, pulumi.Provider(k8sProvider))

		if err != nil {
			return err
		}

		//TODO Create bucket and give default service account access
		bucket, err := storage.NewBucket(ctx, "my-fine-cluster-bucket", &storage.BucketArgs{
			Name:     pulumi.String("myfinecluster"),
			Location: pulumi.String("US"), // FIXME: make configurable
			Labels: pulumi.StringMap{
				"workspace": pulumi.String("myfinecluster"),
			},
		})
		if err != nil {
			return err
		}

		defaultSA := compute.GetDefaultServiceAccountOutput(ctx, compute.GetDefaultServiceAccountOutputArgs{}, nil)
		if err != nil {
			return err
		}

		_, err = storage.NewBucketIAMMember(ctx, "bucket-role", &storage.BucketIAMMemberArgs{
			Bucket: bucket.Name,
			Role:   pulumi.String("roles/storage.admin"),
			Member: defaultSA.Email().ApplyT(func(s string) string { return "serviceAccount:" + s }).(pulumi.StringOutput), //Might need "serviceaccount:" prefix
		})
		if err != nil {
			return err
		}

		_, err = helm.NewRelease(ctx, "pach-1234", &helm.ReleaseArgs{
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://helm.pachyderm.com"),
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
					"host":    pulumi.String("myfinecluster.clusters-ci.pachyderm.io"), // Dynamic Value
					"tls": pulumi.Map{
						"enabled":    pulumi.Bool(true),
						"secretName": pulumi.String("wildcard-tls"), // Dynamic Value
					},
				},
				"pachd": pulumi.Map{
					"externalService": pulumi.Map{
						"enabled": pulumi.Bool(true),
						//"loadBalancerIP": pulumi.String("34.138.68.211"), //Dynamic Value
						"apiGRPCPort":   pulumi.Int(30650), //Dynamic Value
						"s3GatewayPort": pulumi.Int(30600), //Dynamic Value
					},
					"enterpriseLicenseKey": pulumi.String(""), //TODO
					"storage": pulumi.Map{
						"google": pulumi.Map{
							"bucket": bucket.Name, //Dynamic Value
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

func generateKubeconfig(clusterEndpoint pulumi.StringOutput, clusterName pulumi.StringOutput,
	clusterMasterAuth container.ClusterMasterAuthOutput) pulumi.StringOutput {
	context := pulumi.Sprintf("demo_%s", clusterName)

	return pulumi.Sprintf(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: %s
    server: https://%s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
kind: Config
preferences: {}
users:
- name: %s
  user:
    auth-provider:
      config:
        cmd-args: config config-helper --format=json
        cmd-path: gcloud
        expiry-key: '{.credential.token_expiry}'
        token-key: '{.credential.access_token}'
      name: gcp`,
		clusterMasterAuth.ClusterCaCertificate().Elem(),
		clusterEndpoint, context, context, context, context, context, context)
}
