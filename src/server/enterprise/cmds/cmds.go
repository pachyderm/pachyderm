package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/license"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
)

func newClient(enterprise bool) (*client.APIClient, error) {
	if enterprise {
		return client.NewEnterpriseClientOnUserMachine("user")
	}
	return client.NewOnUserMachine("user")
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

			// Register the localhost as a cluster
			resp, err := c.License.AddCluster(c.Ctx(),
				&license.AddClusterRequest{
					Id:      "localhost",
					Address: "grpc://localhost:653",
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
				pachdUsrAddr = c.GetAddress()
			}

			if pachdAddr == "" {
				pachdAddr = ec.GetAddress()
			}

			if clusterId == "" {
				clusterInfo, inspectErr := c.AdminAPIClient.InspectCluster(c.Ctx(), &types.Empty{})
				if inspectErr != nil {
					return inspectErr
				}
				clusterId = clusterInfo.DeploymentID
			}

			// Register the pachd with the license server
			resp, err := ec.License.AddCluster(ec.Ctx(),
				&license.AddClusterRequest{
					Id:                  id,
					Address:             pachdAddr,
					UserAddress:         pachdUsrAddr,
					ClusterDeploymentId: clusterId,
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
					context.ClusterDeploymentID = cluster.ClusterDeploymentId
					context.PachdAddress = cluster.Address
				} else {
					cfg.V2.Contexts[cluster.Id] = &config.Context{
						ClusterDeploymentID: cluster.ClusterDeploymentId,
						PachdAddress:        cluster.Address,
						Source:              config.ContextSource_IMPORTED,
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
