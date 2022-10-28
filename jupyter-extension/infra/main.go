package main

import (
	"io/ioutil"

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
		currBranch := cfg.Require("branch")
		sha := cfg.Require("sha")

		kubeConfig := stackRef.GetOutput(pulumi.String("kubeconfig"))

		k8sProvider, err := kubernetes.NewProvider(ctx, "k8sprovider", &kubernetes.ProviderArgs{
			Kubeconfig: pulumi.StringOutput(kubeConfig),
		})
		if err != nil {
			return err
		}

		namespace, err := corev1.NewNamespace(ctx, "jh-ns", &corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(currBranch),
				Labels: pulumi.StringMap{
					"needs-ci-tls": pulumi.String("true"), //Uses kubernetes replicator to replicate TLS secret to new NS
				},
			},
		}, pulumi.Provider(k8sProvider))

		if err != nil {
			return err
		}

		//file, err := ioutil.ReadFile("./root.py")
		volumeFile, err := ioutil.ReadFile("./volume.py")

		if err != nil {
			return err
		}
		volumeStr := string(volumeFile)
		//fileStr := string(file)
		nbUserImage := "pachyderm/notebooks-user" + ":" + sha
		_, err = helm.NewRelease(ctx, "jh-release", &helm.ReleaseArgs{
			Timeout:   pulumi.Int(600),
			Namespace: namespace.Metadata.Elem().Name(),
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://jupyterhub.github.io/helm-chart/"),
			},
			Chart: pulumi.String("jupyterhub"),
			Values: pulumi.Map{
				"singleuser": pulumi.Map{
					"defaultUrl": pulumi.String("/lab"),
					"cloudMetadata": pulumi.Map{
						"blockWithIptables": pulumi.Bool(false),
					},
					"profileList": pulumi.MapArray{
						pulumi.Map{
							"display_name": pulumi.String("combined"),
							"description":  pulumi.String("Run mount server in Jupyter container"),
							"slug":         pulumi.String("combined"),
							"default":      pulumi.Bool(true),
							"kubespawner_override": pulumi.Map{
								"image":  pulumi.String(nbUserImage),
								"cmd":    pulumi.String("start-singleuser.sh"),
								"uid":    pulumi.Int(0),
								"fs_gid": pulumi.Int(0),
								"environment": pulumi.Map{
									"GRANT_SUDO":         pulumi.String("yes"),
									"NOTEBOOK_ARGS":      pulumi.String("--allow-root"),
									"JUPYTER_ENABLE_LAB": pulumi.String("yes"),
									"CHOWN_HOME":         pulumi.String("yes"),
									"CHOWN_HOME_OPTS":    pulumi.String("-R"),
								},
								"container_security_context": pulumi.Map{
									"allowPrivilegeEscalation": pulumi.Bool(true),
									"runAsUser":                pulumi.Int(0),
									"privileged":               pulumi.Bool(true),
									"capabilities": pulumi.Map{
										"add": pulumi.StringArray{pulumi.String("SYS_ADMIN")},
									},
								},
							},
						},
						pulumi.Map{

							"display_name": pulumi.String("sidecar"),
							"slug":         pulumi.String("sidecar"),
							"description":  pulumi.String("Run mount server as a sidecar"),
							"kubespawner_override": pulumi.Map{
								"image": pulumi.String(nbUserImage),
								"environment": pulumi.Map{
									"SIDECAR_MODE": pulumi.String("true"),
								},
								"extra_containers": pulumi.MapArray{
									pulumi.Map{
										"name":    pulumi.String("mount-server-manager"),
										"image":   pulumi.String("pachyderm/mount-server:" + sha),
										"command": pulumi.StringArray{pulumi.String("/bin/bash"), pulumi.String("-c"), pulumi.String("mount-server")},
										"securityContext": pulumi.Map{
											"privileged": pulumi.Bool(true),
											"runAsUser":  pulumi.Int(0),
										},
										"volumeMounts": pulumi.MapArray{
											pulumi.Map{
												"name":             pulumi.String("shared-pfs"),
												"mountPath":        pulumi.String("/pfs"),
												"mountPropagation": pulumi.String("Bidirectional"),
											},
										},
									},
								},
							},
						},
					},
				},
				//cull
				"ingress": pulumi.Map{
					"enabled": pulumi.Bool(true),
					"annotations": pulumi.Map{
						"kubernetes.io/ingress.class":              pulumi.String("traefik"),
						"traefik.ingress.kubernetes.io/router.tls": pulumi.String("true"),
					},
					"hosts": pulumi.StringArray{pulumi.String("jh-" + currBranch + ".clusters-ci.pachyderm.io")},
					"tls": pulumi.MapArray{
						pulumi.Map{
							"hosts":      pulumi.StringArray{pulumi.String("jh-" + currBranch + ".clusters-ci.pachyderm.io")},
							"secretName": pulumi.String("wildcard-tls"),
						},
					},
				},
				"prePuller": pulumi.Map{
					"hook": pulumi.Map{
						"enabled": pulumi.Bool(false),
					},
				},
				"hub": pulumi.Map{
					"config": pulumi.Map{
						"Auth0OAuthenticator": pulumi.Map{
							"client_id":          pulumi.String(cfg.Require("client_id")),
							"client_secret":      pulumi.String(cfg.Require("client_secret")),
							"oauth_callback_url": pulumi.String("https://jh-" + currBranch + ".clusters-ci.pachyderm.io/hub/oauth_callback"),
							"scope":              pulumi.StringArray{pulumi.String("openid"), pulumi.String("email")},
							"auth0_subdomain":    pulumi.String(cfg.Require("auth0_domain")),
						},
						"Authenticator": pulumi.Map{
							"auto_login": pulumi.Bool(true),
						},
						"JupyterHub": pulumi.Map{
							"authenticator_class": pulumi.String("auth0"),
						},
					},
					"extraConfig": pulumi.Map{
						//"podRoot": pulumi.String(fileStr),
						"volume": pulumi.String(volumeStr),
					},
				},
				"proxy": pulumi.Map{
					"service": pulumi.Map{
						"type": pulumi.String("ClusterIP"),
					},
				},
				"scheduling": pulumi.Map{
					"userScheduler": pulumi.Map{
						"enabled": pulumi.Bool(false),
					},
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
				"deployTarget": pulumi.String("LOCAL"),
				"global": pulumi.Map{
					"postgresql": pulumi.Map{
						"postgresqlPassword": pulumi.String("Vq90lGAZBA"),
					},
				},
				"pachd": pulumi.Map{
					"annotations": pulumi.Map{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": pulumi.String("true"),
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
