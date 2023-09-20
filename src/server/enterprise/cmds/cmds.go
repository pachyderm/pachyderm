package cmds

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/spf13/cobra"
)

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

// DeactivateCmd returns a cobra.Command to deactivate the enterprise service.
func DeactivateCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	deactivate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deactivate the enterprise service",
		Long:  "Deactivate the enterprise service",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			// Deactivate the enterprise server
			req := &enterprise.DeactivateRequest{}
			if _, err := c.Enterprise.Deactivate(c.Ctx(), req); err != nil {
				return errors.EnsureStack(err)
			}

			return nil
		}),
	}

	return cmdutil.CreateAlias(deactivate, "enterprise deactivate")
}

// RegisterCmd returns a cobra.Command that registers this cluster with a remote Enterprise Server.
func RegisterCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var id, pachdAddr, pachdUsrAddr, enterpriseAddr, clusterId string
	register := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Register the cluster with an enterprise license server",
		Long:  "Register the cluster with an enterprise license server",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, false)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			ec, err := pachctlCfg.NewOnUserMachine(ctx, true)
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
				clusterInfo, inspectErr := c.AdminAPIClient.InspectCluster(c.Ctx(), &admin.InspectClusterRequest{
					ClientVersion: version.Version,
				})
				if inspectErr != nil {
					return errors.Wrapf(inspectErr, "could not inspect cluster")
				}
				clusterId = clusterInfo.DeploymentId
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
				err = errors.Wrapf(err, "could not register with the license server.")
				_, deleteErr := ec.License.DeleteCluster(ec.Ctx(), &license.DeleteClusterRequest{Id: id})
				if deleteErr != nil {
					deleteErr := errors.Wrapf(deleteErr, "also failed to rollback creation of cluster with ID, %v."+
						"To retry enterprise registration, first delete this cluster with 'pachctl license delete-cluster --id %v'.: %v",
						id, id, deleteErr.Error())
					return errors.Wrap(err, deleteErr.Error())
				}
				return err
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
func GetStateCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var isEnterprise bool
	getState := &cobra.Command{
		Short: "Check whether the Pachyderm cluster has enterprise features " +
			"activated",
		Long: "Check whether the Pachyderm cluster has enterprise features " +
			"activated",
		Run: cmdutil.Run(func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, isEnterprise)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			resp, err := c.Enterprise.GetState(c.Ctx(), &enterprise.GetStateRequest{})
			if err != nil {
				return errors.EnsureStack(err)
			}
			if resp.State == enterprise.State_NONE {
				fmt.Println("No Pachyderm Enterprise token was found")
				return nil
			}
			ts := resp.Info.Expires.AsTime()
			fmt.Printf("Pachyderm Enterprise token state: %s\nExpiration: %s\n",
				resp.State.String(), ts.String())
			return nil
		}),
	}
	getState.PersistentFlags().BoolVar(&isEnterprise, "enterprise", false, "Activate auth on the active enterprise context")
	return cmdutil.CreateAlias(getState, "enterprise get-state")
}

func SyncContextsCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	syncContexts := &cobra.Command{
		Short: "Pull all available Pachyderm Cluster contexts into your pachctl config",
		Long:  "Pull all available Pachyderm Cluster contexts into your pachctl config",
		Run: cmdutil.Run(func(args []string) error {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}

			ec, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer ec.Close()

			resp, err := ec.License.ListUserClusters(ec.Ctx(), &license.ListUserClustersRequest{})
			if err != nil {
				return errors.EnsureStack(err)
			}

			// update the pach_address of all existing contexts, and add the rest as well.
			for _, cluster := range resp.Clusters {
				if context, ok := cfg.V2.Contexts[cluster.Id]; ok {
					// reset the session token if the context is pointing to a new cluster deployment
					if cluster.ClusterDeploymentId != context.ClusterDeploymentId {
						context.ClusterDeploymentId = cluster.ClusterDeploymentId
						context.SessionToken = ""
					}
					context.PachdAddress = cluster.Address
					context.EnterpriseServer = cluster.EnterpriseServer
				} else {
					cfg.V2.Contexts[cluster.Id] = &config.Context{
						ClusterDeploymentId: cluster.ClusterDeploymentId,
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

// HeartbeatCmd triggers an explicit heartbeat to the license server
func HeartbeatCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var isEnterprise bool
	heartbeat := &cobra.Command{
		Short: "Sync the enterprise state with the license server immediately.",
		Long:  "Sync the enterprise state with the license server immediately.",
		Run: cmdutil.Run(func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, isEnterprise)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			_, err = c.Enterprise.Heartbeat(c.Ctx(), &enterprise.HeartbeatRequest{})
			if err != nil {
				return errors.Wrapf(err, "could not sync with license server")
			}
			return nil
		}),
	}
	heartbeat.PersistentFlags().BoolVar(&isEnterprise, "enterprise", false, "Make the enterprise server refresh its state")
	return cmdutil.CreateAlias(heartbeat, "enterprise heartbeat")
}

// PauseCmd pauses the cluster.
func PauseCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	pause := &cobra.Command{
		Short: "Pause the cluster.",
		Long:  "Pause the cluster.",
		Run: cmdutil.Run(func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			_, err = c.Enterprise.Pause(c.Ctx(), &enterprise.PauseRequest{})
			if err != nil {
				return errors.Wrapf(err, "could not pause cluster")
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(pause, "enterprise pause")
}

// UnpauseCmd pauses the cluster.
func UnpauseCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	unpause := &cobra.Command{
		Short: "Unpause the cluster.",
		Long:  "Unpause the cluster.",
		Run: cmdutil.Run(func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			_, err = c.Enterprise.Unpause(c.Ctx(), &enterprise.UnpauseRequest{})
			if err != nil {
				return errors.Wrapf(err, "could not unpause cluster")
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(unpause, "enterprise unpause")
}

// PauseStatusCmd returns the pause status of the cluster: unpaused; partially
// paused; or completely paused.
func PauseStatusCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	pauseStatus := &cobra.Command{
		Short: "Get the pause status of the cluster.",
		Long:  "Get the pause the cluster: normal, partially-paused or paused.",
		Run: cmdutil.Run(func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			resp, err := c.Enterprise.PauseStatus(c.Ctx(), &enterprise.PauseStatusRequest{})
			if err != nil {
				return errors.Wrapf(err, "could not get pause status")
			}
			switch resp.Status {
			case enterprise.PauseStatusResponse_UNPAUSED:
				fmt.Println("unpaused")
			case enterprise.PauseStatusResponse_PARTIALLY_PAUSED:
				fmt.Println("partially-paused")
			case enterprise.PauseStatusResponse_PAUSED:
				fmt.Println("paused")
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(pauseStatus, "enterprise pause-status")
}

// Cmds returns pachctl commands related to Pachyderm Enterprise
func Cmds(mainCtx context.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	enterprise := &cobra.Command{
		Short: "Enterprise commands enable Pachyderm Enterprise features",
		Long:  "Enterprise commands enable Pachyderm Enterprise features",
	}
	commands = append(commands, cmdutil.CreateAlias(enterprise, "enterprise"))

	commands = append(commands, RegisterCmd(mainCtx, pachctlCfg))
	commands = append(commands, DeactivateCmd(mainCtx, pachctlCfg))
	commands = append(commands, GetStateCmd(mainCtx, pachctlCfg))
	commands = append(commands, SyncContextsCmd(mainCtx, pachctlCfg))
	commands = append(commands, HeartbeatCmd(mainCtx, pachctlCfg))
	commands = append(commands, PauseCmd(mainCtx, pachctlCfg))
	commands = append(commands, UnpauseCmd(mainCtx, pachctlCfg))
	commands = append(commands, PauseStatusCmd(mainCtx, pachctlCfg))

	return commands
}
