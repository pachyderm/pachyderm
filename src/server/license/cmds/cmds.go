package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/license"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
)

// ActivateCmd returns a cobra.Command to activate the license service,
// register the current pachd and activate enterprise features.
// This always runs against the current enterprise context, and can
// be used to activate a single-node pachd deployment or the enterprise
// server in a multi-node deployment.
func ActivateCmd() *cobra.Command {
	var onlyActivate bool
	activate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Activate the license server with an activation code",
		Long:  "Activate the license server with an activation code",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			key, err := cmdutil.ReadPassword("Enterprise key: ")
			if err != nil {
				return errors.Wrapf(err, "could not read enterprise key")
			}

			// Activate the license server
			req := &license.ActivateRequest{
				ActivationCode: key,
			}
			c := env.EnterpriseClient("user")
			if _, err := c.License.Activate(c.Ctx(), req); err != nil {
				return err
			}

			if onlyActivate {
				return nil
			}

			// inspect the activated cluster for its Deployment Id
			clusterInfo, inspectErr := c.AdminAPIClient.InspectCluster(c.Ctx(), &types.Empty{})
			if inspectErr != nil {
				return errors.Wrapf(err, "could not inspect cluster")
			}

			// Register the localhost as a cluster
			resp, err := c.License.AddCluster(c.Ctx(),
				&license.AddClusterRequest{
					Id:                  "localhost",
					Address:             "grpc://localhost:1653",
					UserAddress:         "grpc://localhost:1653",
					ClusterDeploymentId: clusterInfo.DeploymentID,
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
func AddClusterCmd() *cobra.Command {
	var id, address, secret string
	addCluster := &cobra.Command{
		Short: "Register a new cluster with the license server.",
		Long:  "Register a new cluster with the license server.",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {

			c := env.EnterpriseClient("user")
			resp, err := c.License.AddCluster(c.Ctx(), &license.AddClusterRequest{
				Id:      id,
				Address: address,
				Secret:  secret,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(env.Stdout(), "Shared secret: %v\n", resp.Secret)

			return nil
		}),
	}
	addCluster.PersistentFlags().StringVar(&id, "id", "", `The id for the cluster to register`)
	addCluster.PersistentFlags().StringVar(&address, "address", "", `The host and port where the cluster can be reached`)
	addCluster.PersistentFlags().StringVar(&secret, "secret", "", `The shared secret to use to authenticate this cluster`)
	return cmdutil.CreateAlias(addCluster, "license add-cluster")
}

// UpdateClusterCmd returns a cobra.Command to register a cluster with the license server
func UpdateClusterCmd() *cobra.Command {
	var id, address, userAddress, clusterDeploymentId string
	updateCluster := &cobra.Command{
		Short: "Update an existing cluster registered with the license server.",
		Long:  "Update an existing cluster registered with the license server.",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			_, err := c.License.UpdateCluster(c.Ctx(), &license.UpdateClusterRequest{
				Id:                  id,
				Address:             address,
				UserAddress:         userAddress,
				ClusterDeploymentId: clusterDeploymentId,
			})
			return err
		}),
	}
	updateCluster.PersistentFlags().StringVar(&id, "id", "", `The id for the cluster to update`)
	updateCluster.PersistentFlags().StringVar(&address, "address", "", `The host and port where the cluster can be reached by the enterprise server`)
	updateCluster.PersistentFlags().StringVar(&userAddress, "user-address", "", `The host and port where the cluster can be reached by a user`)
	updateCluster.PersistentFlags().StringVar(&clusterDeploymentId, "cluster-deployment-id", "", `The deployment id of the updated cluster`)
	return cmdutil.CreateAlias(updateCluster, "license update-cluster")
}

// DeleteClusterCmd returns a cobra.Command to delete a cluster from the license server
func DeleteClusterCmd() *cobra.Command {
	var id string
	deleteCluster := &cobra.Command{
		Short: "Delete a cluster registered with the license server.",
		Long:  "Delete a cluster registered with the license server.",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			_, err := c.License.DeleteCluster(c.Ctx(), &license.DeleteClusterRequest{
				Id: id,
			})
			return err
		}),
	}
	deleteCluster.PersistentFlags().StringVar(&id, "id", "", `The id for the cluster to delete`)
	return cmdutil.CreateAlias(deleteCluster, "license delete-cluster")
}

// ListClustersCmd returns a cobra.Command to list clusters registered with the license server
func ListClustersCmd() *cobra.Command {
	listClusters := &cobra.Command{
		Short: "List clusters registered with the license server.",
		Long:  "List clusters registered with the license server.",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			resp, err := c.License.ListClusters(c.Ctx(), &license.ListClustersRequest{})
			if err != nil {
				return err
			}

			for _, cluster := range resp.Clusters {
				fmt.Fprintf(env.Stdout(), "id: %v\naddress: %v\nversion: %v\nauth_enabled: %v\nlast_heartbeat: %v\n---\n", cluster.Id, cluster.Address, cluster.Version, cluster.AuthEnabled, cluster.LastHeartbeat)
			}

			return nil
		}),
	}
	return cmdutil.CreateAlias(listClusters, "license list-clusters")
}

// DeleteAllCmd returns a cobra.Command to disable enterprise features and
// clear the configuration of the license service.
func DeleteAllCmd() *cobra.Command {
	activate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Delete all data from the license server",
		Long:  "Delete all data from the license server",
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			if _, err := c.License.DeleteAll(c.Ctx(), &license.DeleteAllRequest{}); err != nil {
				return err
			}
			fmt.Fprintf(env.Stdout(), "All data deleted from license server.")
			return nil
		}),
	}

	return cmdutil.CreateAlias(activate, "license delete-all")
}

// GetStateCmd returns a cobra.Command to get the state of the license service.
func GetStateCmd() *cobra.Command {
	getState := &cobra.Command{
		Short: "Get the configuration of the license service.",
		Long:  "Get the configuration of the license service.",
		RunE: cmdutil.Run(func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			resp, err := c.License.GetActivationCode(c.Ctx(), &license.GetActivationCodeRequest{})
			if err != nil {
				return err
			}
			if resp.State == enterprise.State_NONE {
				fmt.Fprintln(env.Stdout(), "No Pachyderm Enterprise license is configured")
				return nil
			}
			ts, err := types.TimestampFromProto(resp.Info.Expires)
			if err != nil {
				return errors.Wrapf(err, "expiration timestamp could not be parsed")
			}
			fmt.Fprintf(env.Stdout(), "Pachyderm Enterprise token state: %s\nExpiration: %s\nLicense: %s\n",
				resp.State.String(), ts.String(), resp.ActivationCode)
			return nil
		}),
	}
	return cmdutil.CreateAlias(getState, "license get-state")
}

// Cmds returns pachctl commands related to Pachyderm Enterprise
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	enterprise := &cobra.Command{
		Short: "License commmands manage the Enterprise License service",
		Long:  "License commands manage the Enterprise License service",
	}
	commands = append(commands, cmdutil.CreateAlias(enterprise, "license"))
	commands = append(commands, ActivateCmd())
	commands = append(commands, AddClusterCmd())
	commands = append(commands, UpdateClusterCmd())
	commands = append(commands, DeleteClusterCmd())
	commands = append(commands, ListClustersCmd())
	commands = append(commands, DeleteAllCmd())
	commands = append(commands, GetStateCmd())

	return commands
}
