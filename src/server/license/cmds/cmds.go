// Package cmds implements pachctl commands for the license service.
package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/spf13/cobra"
)

// ActivateCmd returns a cobra.Command to activate the license service,
// register the current pachd and activate enterprise features.
// This always runs against the current enterprise context, and can
// be used to activate a single-node pachd deployment or the enterprise
// server in a multi-node deployment.
func ActivateCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	var onlyActivate bool
	activate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Activate the license server with an activation code",
		Long:  "This command activates Enterprise Server with an activation code.",
		Example: "\t- {{alias}}\n" +
			"\t- {{alias}} --no-register\n",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			key, err := cmdutil.ReadPassword("Enterprise key: ")
			if err != nil {
				return errors.Wrapf(err, "could not read enterprise key")
			}

			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer errors.Close(&retErr, c, "close client")

			// Activate the license server
			req := &license.ActivateRequest{
				ActivationCode: key,
			}
			if _, err := c.License.Activate(c.Ctx(), req); err != nil {
				return errors.EnsureStack(err)
			}
			if onlyActivate {
				return nil
			}

			clusterInfo, ok := c.ClusterInfo()
			if !ok {
				return errors.New("internal: no cluster info in pach client?")
			}

			// Register the localhost as a cluster
			resp, err := c.License.AddCluster(c.Ctx(),
				&license.AddClusterRequest{
					Id:                  "localhost",
					Address:             "grpc://localhost:1653",
					UserAddress:         "grpc://localhost:1653",
					ClusterDeploymentId: clusterInfo.DeploymentId,
					EnterpriseServer:    true,
				})
			if err != nil {
				return errors.Wrapf(err, "could not register pachd with the license service")
			}

			// activate the Enterprise service
			_, err = c.Enterprise.Activate(c.Ctx(),
				&enterprise.ActivateRequest{
					Id:            "localhost",
					Secret:        resp.Secret,
					LicenseServer: "grpc://localhost:1653",
				})
			if err != nil {
				return errors.Wrapf(err, "could not activate the enterprise service")
			}

			return nil

		}),
	}
	activate.PersistentFlags().BoolVar(&onlyActivate, "no-register", false, "Activate auth on the active enterprise context")
	return cmdutil.CreateAlias(activate, "license activate")
}

// AddClusterCmd returns a cobra.Command to register a cluster with the license server
func AddClusterCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	var id, address, secret string
	addCluster := &cobra.Command{
		Short:   "Register a new cluster with the license server.",
		Long:    "This command registers a new cluster with Enterprise Server.",
		Example: "\t- {{alias}} --id=my-cluster --address=grpc://my-cluster:1653 --secret=secret\n",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer errors.Close(&retErr, c, "close client")

			resp, err := c.License.AddCluster(c.Ctx(), &license.AddClusterRequest{
				Id:      id,
				Address: address,
				Secret:  secret,
			})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			fmt.Printf("Shared secret: %v\n", resp.Secret)

			return nil
		}),
	}
	addCluster.PersistentFlags().StringVar(&id, "id", "", `Set the ID for the cluster to register.`)
	addCluster.PersistentFlags().StringVar(&address, "address", "", `Set the host and port where the cluster can be reached.`)
	addCluster.PersistentFlags().StringVar(&secret, "secret", "", `Set the shared secret to use to authenticate this cluster.`)
	return cmdutil.CreateAlias(addCluster, "license add-cluster")
}

// UpdateClusterCmd returns a cobra.Command to register a cluster with the license server
func UpdateClusterCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	var id, address, userAddress, clusterDeploymentId string
	updateCluster := &cobra.Command{
		Short: "Update an existing cluster registered with the license server.",
		Long:  "This command updates an existing cluster registered with Enterprise Server.",
		Example: "\t- {{alias}} --id=my-cluster --address=grpc://my-cluster:1653 \n" +
			"\t- {{alias}} --id=my-cluster --user-address=grpc://my-cluster:1653\n" +
			"\t- {{alias}} --id=my-cluster --cluster-deployment-id=1234\n",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer errors.Close(&retErr, c, "close client")

			_, err = c.License.UpdateCluster(c.Ctx(), &license.UpdateClusterRequest{
				Id:                  id,
				Address:             address,
				UserAddress:         userAddress,
				ClusterDeploymentId: clusterDeploymentId,
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	updateCluster.PersistentFlags().StringVar(&id, "id", "", `Set the ID for the cluster to update.`)
	updateCluster.PersistentFlags().StringVar(&address, "address", "", `Set the host and port where the cluster can be reached by the enterprise server.`)
	updateCluster.PersistentFlags().StringVar(&userAddress, "user-address", "", `Set the host and port where the cluster can be reached by a user.`)
	updateCluster.PersistentFlags().StringVar(&clusterDeploymentId, "cluster-deployment-id", "", `Set the deployment ID of the updated cluster.`)
	return cmdutil.CreateAlias(updateCluster, "license update-cluster")
}

// DeleteClusterCmd returns a cobra.Command to delete a cluster from the license server
func DeleteClusterCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	var id string
	deleteCluster := &cobra.Command{
		Short:   "Delete a cluster registered with the license server.",
		Long:    "This command deletes a cluster registered with Enterprise Server.",
		Example: "\t- {{alias}} --id=my-cluster\n",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer errors.Close(&retErr, c, "close client")

			_, err = c.License.DeleteCluster(c.Ctx(), &license.DeleteClusterRequest{
				Id: id,
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	deleteCluster.PersistentFlags().StringVar(&id, "id", "", `Set the ID for the cluster to delete.`)
	return cmdutil.CreateAlias(deleteCluster, "license delete-cluster")
}

// ListClustersCmd returns a cobra.Command to list clusters registered with the license server
func ListClustersCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	listClusters := &cobra.Command{
		Short: "List clusters registered with the license server.",
		Long:  "This command lists clusters registered with Enterprise Server.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer errors.Close(&retErr, c, "close client")

			resp, err := c.License.ListClusters(c.Ctx(), &license.ListClustersRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			for _, cluster := range resp.Clusters {
				fmt.Printf("id: %v\naddress: %v\nversion: %v\nauth_enabled: %v\nlast_heartbeat: %v\n---\n", cluster.Id, cluster.Address, cluster.Version, cluster.AuthEnabled, cluster.LastHeartbeat)
			}

			return nil
		}),
	}
	return cmdutil.CreateAlias(listClusters, "license list-clusters")
}

// DeleteAllCmd returns a cobra.Command to disable enterprise features and
// clear the configuration of the license service.
func DeleteAllCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	activate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Delete all data from the license server",
		Long:  "This command deletes all data from Enterprise Server.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer errors.Close(&retErr, c, "close client")

			if _, err := c.License.DeleteAll(c.Ctx(), &license.DeleteAllRequest{}); err != nil {
				return errors.EnsureStack(err)
			}
			fmt.Printf("All data deleted from license server.")
			return nil
		}),
	}

	return cmdutil.CreateAlias(activate, "license delete-all")
}

// GetStateCmd returns a cobra.Command to get the state of the license service.
func GetStateCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	getState := &cobra.Command{
		Short: "Get the configuration of the license service.",
		Long:  "This command returns the configuration of the Enterprise Server.",
		Run: cmdutil.Run(func(cmd *cobra.Command, args []string) (retErr error) {
			ctx := cmd.Context()
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer errors.Close(&retErr, c, "close client")

			resp, err := c.License.GetActivationCode(c.Ctx(), &license.GetActivationCodeRequest{})
			if err != nil {
				return errors.EnsureStack(err)
			}
			if resp.State == enterprise.State_NONE {
				fmt.Println("No Pachyderm Enterprise license is configured")
				return nil
			}
			ts := resp.GetInfo().GetExpires().AsTime()
			fmt.Printf("Pachyderm Enterprise token state: %s\nExpiration: %s\nLicense: %s\n",
				resp.State.String(), ts.String(), resp.ActivationCode)
			return nil
		}),
	}
	return cmdutil.CreateAlias(getState, "license get-state")
}

// Cmds returns pachctl commands related to Pachyderm Enterprise
func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	enterprise := &cobra.Command{
		Short: "License commmands manage the Enterprise License service",
		Long:  "License commands manage the Enterprise License service",
	}
	commands = append(commands, cmdutil.CreateAlias(enterprise, "license"))
	commands = append(commands, ActivateCmd(pachctlCfg))
	commands = append(commands, AddClusterCmd(pachctlCfg))
	commands = append(commands, UpdateClusterCmd(pachctlCfg))
	commands = append(commands, DeleteClusterCmd(pachctlCfg))
	commands = append(commands, ListClustersCmd(pachctlCfg))
	commands = append(commands, DeleteAllCmd(pachctlCfg))
	commands = append(commands, GetStateCmd(pachctlCfg))

	return commands
}
