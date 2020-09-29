package cmds

import (
	"bytes"
	"fmt"
	"os"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/helm"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/spf13/cobra"
)

const (
	ideNotes = `
Thanks for installing the Pachyderm IDE!

It may take a few minutes for all of the pods to spin up. If you have kubectl
access, you can check progress with:

  kubectl get pod -l release=pachyderm-ide

Once all of the pods are in the 'Ready' status, you can access the IDE in the
following manners:

* If you're on docker for mac, it should be accessible on 'localhost'.
* If you're on minikube, run 'minikube service proxy-public --url' -- one or
  both of the URLs printed should reach the IDE.
* If you're on a cloud deployment, use the external IP of
  'kubectl get service proxy-public'.

For more information about the Pachyderm IDE, see these resources:

* Our how-tos: https://docs.pachyderm.com/latest/how-tos/use-pachyderm-ide/
* The Z2JH docs, which the IDE builds off of:
  https://zero-to-jupyterhub.readthedocs.io/en/latest/
`

	defaultIDEHubImage  = "pachyderm/ide-hub"
	defaultIDEUserImage = "pachyderm/ide-user"

	defaultIDEVersion      = "1.1.0"
	defaultIDEChartVersion = "0.9.1" // see https://jupyterhub.github.io/helm-chart/
)

func CreateDeployIDECommand() *cobra.Command {
	var lbTLSHost string
	var lbTLSEmail string
	var dryRun bool
	var outputFormat string
	var jupyterhubChartVersion string
	var hubImage string
	var userImage string
	deployIDE := &cobra.Command{
		Short: "Deploy the Pachyderm IDE.",
		Long:  "Deploy a JupyterHub-based IDE alongside the Pachyderm cluster.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			cfg, err := config.Read(false)
			if err != nil {
				return err
			}
			_, activeContext, err := cfg.ActiveContext(true)
			if err != nil {
				return err
			}

			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "error constructing pachyderm client")
			}
			defer c.Close()

			enterpriseResp, err := c.Enterprise.GetState(c.Ctx(), &enterprise.GetStateRequest{})
			if err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not get Enterprise status")
			}

			if enterpriseResp.State != enterprise.State_ACTIVE {
				return errors.New("Pachyderm Enterprise must be enabled to use this feature")
			}

			authActive, err := c.IsAuthActive()
			if err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not check whether auth is active")
			}
			if !authActive {
				return errors.New("Pachyderm auth must be enabled to use this feature")
			}

			whoamiResp, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
			if err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not get the current logged in user")
			}

			authTokenResp, err := c.GetAuthToken(c.Ctx(), &auth.GetAuthTokenRequest{
				Subject: whoamiResp.Username,
			})
			if err != nil {
				return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not get an auth token")
			}

			if jupyterhubChartVersion == "" {
				jupyterhubChartVersion = getCompatibleVersion("jupyterhub", "/jupyterhub", defaultIDEChartVersion)
			}
			if hubImage == "" || userImage == "" {
				ideVersion := getCompatibleVersion("ide", "/ide", defaultIDEVersion)
				if hubImage == "" {
					hubImage = fmt.Sprintf("%s:%s", defaultIDEHubImage, ideVersion)
				}
				if userImage == "" {
					userImage = fmt.Sprintf("%s:%s", defaultIDEUserImage, ideVersion)
				}
			}

			hubImageName, hubImageTag := docker.ParseRepositoryTag(hubImage)
			userImageName, userImageTag := docker.ParseRepositoryTag(userImage)

			values := map[string]interface{}{
				"hub": map[string]interface{}{
					"image": map[string]interface{}{
						"name": hubImageName,
						"tag":  hubImageTag,
					},
					"extraConfig": map[string]interface{}{
						"templates": "c.JupyterHub.template_paths = ['/app/templates']",
					},
				},
				"singleuser": map[string]interface{}{
					"image": map[string]interface{}{
						"name": userImageName,
						"tag":  userImageTag,
					},
					"defaultUrl": "/lab",
				},
				"auth": map[string]interface{}{
					"state": map[string]interface{}{
						"enabled":   true,
						"cryptoKey": generateSecureToken(16),
					},
					"type": "custom",
					"custom": map[string]interface{}{
						"className": "pachyderm_authenticator.PachydermAuthenticator",
						"config": map[string]interface{}{
							"pach_auth_token": authTokenResp.Token,
						},
					},
					"admin": map[string]interface{}{
						"users": []string{whoamiResp.Username},
					},
				},
				"proxy": map[string]interface{}{
					"secretToken": generateSecureToken(16),
				},
			}

			if lbTLSHost != "" && lbTLSEmail != "" {
				values["https"] = map[string]interface{}{
					"hosts": []string{lbTLSHost},
					"letsencrypt": map[string]interface{}{
						"contactEmail": lbTLSEmail,
					},
				}
			}

			if dryRun {
				var buf bytes.Buffer
				enc := encoder(outputFormat, &buf)
				if err = enc.Encode(values); err != nil {
					return err
				}
				_, err = os.Stdout.Write(buf.Bytes())
				return err
			}

			_, err = helm.Deploy(
				activeContext,
				"jupyterhub",
				"https://jupyterhub.github.io/helm-chart/",
				"pachyderm-ide",
				"jupyterhub/jupyterhub",
				jupyterhubChartVersion,
				values,
			)
			if err != nil {
				return errors.Wrapf(err, "failed to deploy Pachyderm IDE")
			}

			fmt.Println(ideNotes)
			return nil
		}),
	}
	deployIDE.Flags().StringVar(&lbTLSHost, "lb-tls-host", "", "Hostname for minting a Let's Encrypt TLS cert on the load balancer")
	deployIDE.Flags().StringVar(&lbTLSEmail, "lb-tls-email", "", "Contact email for minting a Let's Encrypt TLS cert on the load balancer")
	deployIDE.Flags().BoolVar(&dryRun, "dry-run", false, "Don't actually deploy, instead just print the Helm config.")
	deployIDE.Flags().StringVarP(&outputFormat, "output", "o", "json", "Output format. One of: json|yaml")
	deployIDE.Flags().StringVar(&jupyterhubChartVersion, "jupyterhub-chart-version", "", "Version of the underlying Zero to JupyterHub with Kubernetes helm chart to use. By default this value is automatically derived.")
	deployIDE.Flags().StringVar(&hubImage, "hub-image", "", "Image for IDE hub. By default this value is automatically derived.")
	deployIDE.Flags().StringVar(&userImage, "user-image", "", "Image for IDE user environments. By default this value is automatically derived.")

	return deployIDE
}
