package cmds

import (
	"bytes"
	"fmt"
	"os"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/helm"
	"github.com/spf13/cobra"
)

const (
	defaultIDEHubImage  = "pachyderm/ide-hub"
	defaultIDEUserImage = "pachyderm/ide-user"

	defaultIDEVersion      = "2.0.0-a2"
	defaultIDEChartVersion = "0.9.1" // see https://jupyterhub.github.io/helm-chart/
	ideNotes               = `
Thanks for installing the Pachyderm IDE!

It may take a few minutes for all of the pods to spin up. If you have kubectl
access, you can check progress with:

	kubectl get pod -l release=pachyderm-ide

Once all of the pods are in the 'Ready' status, you can access the IDE
by running 'pachctl port-forward' and visiting 'localhost:30659'

For more information about the Pachyderm IDE, see these resources:

* Our how-tos: https://docs.pachyderm.com/latest/how-tos/use-pachyderm-ide/
* The Z2JH docs, which the IDE builds off of:
https://zero-to-jupyterhub.readthedocs.io/en/latest/
`
)

func createDeployIDECmd() *cobra.Command {
	var lbTLSHost string
	var lbTLSEmail string
	var dryRun bool
	var outputFormat string
	var jupyterhubChartVersion string
	var hubImage string
	var userImage string
	var clientID string
	var internalIDAddr, idAddr, ideAddr string
	deployIDE := &cobra.Command{
		Short: "Deploy the Pachyderm IDE.",
		Long:  "Deploy a JupyterHub-based IDE alongside the Pachyderm cluster.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			cfg, err := config.Read(false, false)
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

			pachdClientID, clientSecret, err := addOIDCClient(c, clientID, ideAddr+"/hub/oauth_callback")
			if err != nil {
				return err
			}

			hubImageName, hubImageTag := docker.ParseRepositoryTag(hubImage)
			userImageName, userImageTag := docker.ParseRepositoryTag(userImage)

			values := map[string]interface{}{
				"hub": map[string]interface{}{
					"image": map[string]interface{}{
						"name": hubImageName,
						"tag":  hubImageTag,
					},
					"extraEnv": map[string]interface{}{
						"OAUTH2_AUTHORIZE_URL": idAddr + "/auth",
						"OAUTH2_TOKEN_URL":     internalIDAddr + "/token",
						"OAUTH_CALLBACK_URL":   ideAddr + "/hub/oauth_callback",
					},
				},
				"singleuser": map[string]interface{}{
					"image": map[string]interface{}{
						"name": userImageName,
						"tag":  userImageTag,
					},
					"defaultUrl": "/lab",
				},
				"proxy": map[string]interface{}{
					"labels": map[string]interface{}{
						"suite": "pachyderm",
					},
					"secretToken": generateSecureToken(16),
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
							"enable_auth_state": true,
							"client_id":         clientID,
							"client_secret":     clientSecret,
							"token_url":         internalIDAddr + "/token",
							"userdata_url":      internalIDAddr + "/userinfo",
							"scope":             []string{"openid", "email", "audience:server:client_id:" + pachdClientID},
							"username_key":      "email",
						},
					},
					"admin": map[string]interface{}{
						"users": []string{whoamiResp.Username},
					},
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
	deployIDE.Flags().StringVar(&clientID, "client-id", "ide", "The OIDC client ID for the IDE.")
	deployIDE.Flags().StringVar(&internalIDAddr, "internal-id-server", "http://pachd:658", "The web address where the identity server can be reached from within the cluster.")
	deployIDE.Flags().StringVar(&idAddr, "id-server", "http://localhost:30658", "The web address where the identity server can be reached from the client machine.")
	deployIDE.Flags().StringVar(&ideAddr, "ide-address", "http://localhost:30659", "The web address where the IDE can be reached from the client machine.")

	return deployIDE
}
