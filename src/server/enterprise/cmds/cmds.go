package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/license"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"

	"github.com/pkg/browser"
)

func newClient(enterprise bool) (*client.APIClient, error) {
	if enterprise {
		c, err := client.NewEnterpriseClientOnUserMachine("user")
		if err != nil {
			return nil, err
		}
		fmt.Printf("Using enterprise context: %v\n", c.ClientContextName())
		return c, nil
	}
	return client.NewOnUserMachine("user")
}

func getIsActiveContextEnterpriseServer() (bool, error) {
	cfg, err := config.Read(false, true)
	if err != nil {
		return false, errors.Wrapf(err, "could not read config")
	}
	_, ctx, err := cfg.ActiveEnterpriseContext(true)
	if err != nil {
		return false, errors.Wrapf(err, "could not retrieve the enterprise context from the config")
	}
	return ctx.EnterpriseServer, nil
}

func requestOIDCPrivilegedLogin(c *client.APIClient, openBrowser bool) (string, error) {
	var authURL string
	loginInfo, err := c.GetOIDCLogin(c.Ctx(), &auth.GetOIDCLoginRequest{LoginType: auth.OIDCLoginType_PRIVILEDGED})
	if err != nil {
		return "", err
	}
	authURL = loginInfo.LoginURL
	state := loginInfo.State

	// print the prepared URL and promp the user to click on it
	fmt.Println("You will momentarily be directed to your IdP and asked to authorize Pachyderm's " +
		"login app on your IdP.\n\nPaste the following URL into a browser if not automatically redirected:\n\n" +
		authURL + "\n\n" +
		"")

	if openBrowser {
		err = browser.OpenURL(authURL)
		if err != nil {
			return "", err
		}
	}

	return state, nil
}

// ActivateCmd returns a cobra.Command to activate the license service,
// register the current pachd and activate enterprise features.
// This always runs against the current enterprise context, and can
// be used to activate a single-node pachd deployment or the enterprise
// server in a multi-node deployment.
func ActivateCmd() *cobra.Command {
	activate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Activate the license server and enable enterprise features",
		Long:  "Activate the license server and enable enterprise features",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			key, err := cmdutil.ReadPassword("Enterprise key: ")
			if err != nil {
				return errors.Wrapf(err, "could not read enterprise key")
			}

			c, err := newClient(true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			// Activate the license server
			req := &license.ActivateRequest{
				ActivationCode: key,
			}
			if _, err := c.License.Activate(c.Ctx(), req); err != nil {
				return err
			}

			// inspect the activated cluster for its Deployment Id
			clusterInfo, inspectErr := c.AdminAPIClient.InspectCluster(c.Ctx(), &types.Empty{})
			if inspectErr != nil {
				return errors.Wrapf(inspectErr, "could not inspect cluster")
			}

			// inspect the active context to determine whether its pointing at an enterprise server
			enterpriseServer, err := getIsActiveContextEnterpriseServer()
			if err != nil {
				return err
			}

			// Register the localhost as a cluster
			resp, err := c.License.AddCluster(c.Ctx(),
				&license.AddClusterRequest{
					Id:                  "localhost",
					Address:             "grpc://localhost:653",
					UserAddress:         "grpc://localhost:653",
					ClusterDeploymentId: clusterInfo.DeploymentID,
					EnterpriseServer:    enterpriseServer,
				})
			if err != nil {
				return errors.Wrapf(err, "could not register pachd with the license service")
			}

			// activate the Enterprise service
			_, err = c.Enterprise.Activate(c.Ctx(),
				&enterprise.ActivateRequest{
					Id:            "localhost",
					Secret:        resp.Secret,
					LicenseServer: "grpc://localhost:653",
				})
			if err != nil {
				return errors.Wrapf(err, "could not activate the enterprise service")
			}

			return nil
		}),
	}

	return cmdutil.CreateAlias(activate, "enterprise activate")
}

// DeactivateCmd returns a cobra.Command to deactivate the enterprise service.
func DeactivateCmd() *cobra.Command {
	deactivate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deactivate the enterprise service",
		Long:  "Deactivate the enterprise service",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			// Deactivate the enterprise server
			req := &enterprise.DeactivateRequest{}
			if _, err := c.Enterprise.Deactivate(c.Ctx(), req); err != nil {
				return err
			}

			return nil
		}),
	}

	return cmdutil.CreateAlias(deactivate, "enterprise deactivate")
}

// RegisterCmd returns a cobra.Command that registers this cluster with a remote Enterprise Server.
func RegisterCmd() *cobra.Command {
	var id, pachdAddr, pachdUsrAddr, enterpriseAddr, clusterId string
	register := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Register the cluster with an enterprise license server",
		Long:  "Register the cluster with an enterprise license server",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			ec, err := client.NewEnterpriseClientOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer ec.Close()

			if pachdUsrAddr == "" {
				pachdUsrAddr = c.GetAddress().Qualified()
			}

			if pachdAddr == "" {
				pachdAddr = c.GetAddress().Qualified()
			}

			if enterpriseAddr == "" {
				enterpriseAddr = ec.GetAddress().Qualified()
			}

			if clusterId == "" {
				clusterInfo, inspectErr := c.AdminAPIClient.InspectCluster(c.Ctx(), &types.Empty{})
				if inspectErr != nil {
					return errors.Wrapf(inspectErr, "could not inspect cluster")
				}
				clusterId = clusterInfo.DeploymentID
			}

			enterpriseServer, err := getIsActiveContextEnterpriseServer()
			if err != nil {
				return err
			}

			// Register the pachd with the license server
			resp, err := ec.License.AddCluster(ec.Ctx(),
				&license.AddClusterRequest{
					Id:                  id,
					Address:             pachdAddr,
					UserAddress:         pachdUsrAddr,
					ClusterDeploymentId: clusterId,
					EnterpriseServer:    enterpriseServer,
				})
			if err != nil {
				return errors.Wrapf(err, "could not register pachd with the license service")
			}

			// activate the Enterprise service
			_, err = c.Enterprise.Activate(c.Ctx(),
				&enterprise.ActivateRequest{
					Id:            id,
					Secret:        resp.Secret,
					LicenseServer: enterpriseAddr,
				})
			if err != nil {
				return errors.Wrapf(err, "could not register with the license server")
			}

			return nil
		}),
	}
	register.PersistentFlags().StringVar(&id, "id", "", "the id for this cluster")
	register.PersistentFlags().StringVar(&pachdAddr, "pachd-address", "", "the address for the enterprise server to reach this pachd")
	register.PersistentFlags().StringVar(&pachdUsrAddr, "pachd-user-address", "", "the address for a user to reach this pachd")
	register.PersistentFlags().StringVar(&enterpriseAddr, "enterprise-server-address", "", "the address for the pachd to reach the enterprise server")
	register.PersistentFlags().StringVar(&clusterId, "cluster-deployment-id", "", "the deployment id of the cluster being registered")

	return cmdutil.CreateAlias(register, "enterprise register")
}

// GetStateCmd returns a cobra.Command to activate the enterprise features of
// Pachyderm within a Pachyderm cluster. All repos will go from
// publicly-accessible to accessible only by the owner, who can subsequently add
// users
func GetStateCmd() *cobra.Command {
	getState := &cobra.Command{
		Short: "Check whether the Pachyderm cluster has enterprise features " +
			"activated",
		Long: "Check whether the Pachyderm cluster has enterprise features " +
			"activated",
		Run: cmdutil.Run(func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			resp, err := c.Enterprise.GetState(c.Ctx(), &enterprise.GetStateRequest{})
			if err != nil {
				return err
			}
			if resp.State == enterprise.State_NONE {
				fmt.Println("No Pachyderm Enterprise token was found")
				return nil
			}
			ts, err := types.TimestampFromProto(resp.Info.Expires)
			if err != nil {
				return errors.Wrapf(err, "activation request succeeded, but could not "+
					"convert token expiration time to a timestamp")
			}
			fmt.Printf("Pachyderm Enterprise token state: %s\nExpiration: %s\n",
				resp.State.String(), ts.String())
			return nil
		}),
	}
	return cmdutil.CreateAlias(getState, "enterprise get-state")
}

func SyncContextsCmd() *cobra.Command {
	syncContexts := &cobra.Command{
		Short: "Pull all available Pachyderm Cluster contexts into your pachctl config",
		Long:  "Pull all available Pachyderm Cluster contexts into your pachctl config",
		Run: cmdutil.Run(func(args []string) error {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}

			ec, err := client.NewEnterpriseClientOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer ec.Close()

			resp, err := ec.License.ListUserClusters(ec.Ctx(), &license.ListUserClustersRequest{})
			if err != nil {
				return err
			}

			// update the pach_address of all existing contexts, and add the rest as well.
			for _, cluster := range resp.Clusters {
				if context, ok := cfg.V2.Contexts[cluster.Id]; ok {
					// reset the session token if the context is pointing to a new cluster deployment
					if cluster.ClusterDeploymentId != context.ClusterDeploymentID {
						context.ClusterDeploymentID = cluster.ClusterDeploymentId
						context.SessionToken = ""
					}
					context.PachdAddress = cluster.Address
					context.EnterpriseServer = cluster.EnterpriseServer
				} else {
					cfg.V2.Contexts[cluster.Id] = &config.Context{
						ClusterDeploymentID: cluster.ClusterDeploymentId,
						PachdAddress:        cluster.Address,
						Source:              config.ContextSource_IMPORTED,
						EnterpriseServer:    cluster.EnterpriseServer,
					}
				}
			}

			err = cfg.Write()
			if err != nil {
				return err
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(syncContexts, "enterprise sync-contexts")
}

func RevokeUserTokensCmd() *cobra.Command {
	var username string
	revokeUserTokens := &cobra.Command{
		Short: "Revoke a user's auth tokens across all the clusters registered with the enterprise server.",
		Long:  "Revoke a user's auth tokens across all the clusters registered with the enterprise server.",
		Run: cmdutil.Run(func(args []string) error {
			ec, err := client.NewEnterpriseClientOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer ec.Close()

			// first login to get a privileged ID Token
			var idToken string
			noBrowser := false
			if state, err := requestOIDCPrivilegedLogin(ec, !noBrowser); err == nil {
				// Exchange OIDC token for Pachyderm token
				fmt.Println("Retrieving Pachyderm token...")
				resp, authErr := ec.Authenticate(
					ec.Ctx(),
					&auth.AuthenticateRequest{OIDCState: state})
				if authErr != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(authErr),
						"authorization failed (OIDC state token: %q; Pachyderm logs may "+
							"contain more information)",
						// Print state token as it's logged, for easy searching
						fmt.Sprintf("%s.../%d", state[:len(state)/2], len(state)))
				}
				idToken = resp.IdToken
			} else {
				return fmt.Errorf("no authentication providers are configured")
			}

			resp, err := ec.License.RevokeTokensForUserAcrossClusters(ec.Ctx(), &license.RevokeTokensForUserAcrossClustersRequest{Username: username, IdToken: idToken})
			if err != nil {
				return err
			}
			message := ""
			if len(resp.SuccessClusters) > 0 {
				message += "The token was successfully revoked in these clusters:\n"
				for _, c := range resp.SuccessClusters {
					message += fmt.Sprintf("-> %v\n", c)
				}
			}
			if len(resp.FailureClusters) > 0 {
				message += "The user failed to be revoked for these clusters.\n" +
					"Retry running 'enterprise revoke-user-tokens'.\n"
				for _, c := range resp.FailureClusters {
					message += fmt.Sprintf("-> %v. Message: %v\n", c.ReceiverId, c.Err)
				}
			}
			if len(resp.TimedoutClusters) > 0 {
				message += "The following clusters timed out while attempting to revoke the user.\n" +
					"Retry running 'enterprise revoke-user-tokens'.\n"
				for _, c := range resp.TimedoutClusters {
					message += fmt.Sprintf("-> %v\n", c)
				}
			}
			fmt.Println(message)
			return nil
		}),
	}
	revokeUserTokens.PersistentFlags().StringVar(&username, "username", "", "The username of the user whose access should be revoked.")
	return cmdutil.CreateAlias(revokeUserTokens, "enterprise revoke-user-tokens")
}

// Cmds returns pachctl commands related to Pachyderm Enterprise
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	enterprise := &cobra.Command{
		Short: "Enterprise commands enable Pachyderm Enterprise features",
		Long:  "Enterprise commands enable Pachyderm Enterprise features",
	}
	commands = append(commands, cmdutil.CreateAlias(enterprise, "enterprise"))

	commands = append(commands, ActivateCmd())
	commands = append(commands, RegisterCmd())
	commands = append(commands, DeactivateCmd())
	commands = append(commands, GetStateCmd())
	commands = append(commands, SyncContextsCmd())

	return commands
}
