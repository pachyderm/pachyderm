package main

import (
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/projects"
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

		/*engineVersions, err := container.GetEngineVersions(ctx, &container.GetEngineVersionsArgs{}, pulumi.DependsOn([]pulumi.Resource{containerService}))
		if err != nil {
			return err
		}
		masterVersion := engineVersions.LatestMasterVersion*/

		cluster, err := container.NewCluster(ctx, "ci-cluster", &container.ClusterArgs{
			InitialNodeCount: pulumi.Int(2),
			MinMasterVersion: pulumi.String("1.21.5-gke.1302"),
			NodeVersion:      pulumi.String("1.21.5-gke.1302"),
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

		ingressIP, err := compute.NewAddress(ctx, "ingress-ip", nil)
		if err != nil {
			return err
		}
		ctx.Export("ingressIP", ingressIP.Address)

		_, err = helm.NewRelease(ctx, "traefik", &helm.ReleaseArgs{
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://helm.traefik.io/traefik"),
			},
			Chart: pulumi.String("traefik"),
			Values: pulumi.Map{
				"service": pulumi.Map{
					"spec": pulumi.Map{
						"loadBalancerIP": ingressIP.Address,
					},
				},
			},
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

		//TODO Create Cluster Issuer with DNS Validation

		// TODO Create Certificate Resource for wildcard
		/*
			apiVersion: cert-manager.io/v1
			kind: Certificate
			metadata:
			  name: ci-cluster-wildcafd
			spec:
			  secretName: example-com-tls
			  secretTemplate:
			    annotations:
			      replicator.v1.mittwald.de/replicate-to-matching: needs-ci-tls=true
			  issuerRef:
			    name: ca-issuer # Reference above issuer
			    kind: Issuer #Cluster issuer
		*/

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
