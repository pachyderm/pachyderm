package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/license"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
)

// ActivateCmd returns a cobra.Command to activate the license service
func ActivateCmd() *cobra.Command {
	activate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Activate the license server with an activation code",
		Long:  "Activate the license server with an activation code",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			key, err := cmdutil.ReadPassword("Enterprise key: ")
			if err != nil {
				return errors.Wrapf(err, "could not read enterprise key")
			}

			c, err := client.NewOnUserMachine("user")
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

			return nil
		}),
	}

	return cmdutil.CreateAlias(activate, "license activate")
}

// AddClusterCmd returns a cobra.Command to register a cluster with the license server
func AddClusterCmd() *cobra.Command {
	var id, address string
	addCluster := &cobra.Command{
		Short: "Register a new cluster with the license server.",
		Long:  "Register a new cluster with the license server.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}

			resp, err := c.License.AddCluster(c.Ctx(), &license.AddClusterRequest{
				Id:      id,
				Address: address,
			})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			fmt.Printf("Shared secret: %v\n", resp.Secret)

			return nil
		}),
	}
	addCluster.PersistentFlags().StringVar(&id, "id", "", `The id for the cluster to register`)
	addCluster.PersistentFlags().StringVar(&address, "address", "", `The host and port where the cluster can be reached`)
	return cmdutil.CreateAlias(addCluster, "license add-cluster")
}

// UpdateClusterCmd returns a cobra.Command to register a cluster with the license server
func UpdateClusterCmd() *cobra.Command {
	var id, address string
	updateCluster := &cobra.Command{
		Short: "Update an existing cluster registered with the license server.",
		Long:  "Update an existing cluster registered with the license server.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}

			_, err = c.License.UpdateCluster(c.Ctx(), &license.UpdateClusterRequest{
				Id:      id,
				Address: address,
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	updateCluster.PersistentFlags().StringVar(&id, "id", "", `The id for the cluster to update`)
	updateCluster.PersistentFlags().StringVar(&address, "address", "", `The host and port where the cluster can be reached`)
	return cmdutil.CreateAlias(updateCluster, "license update-cluster")
}

// DeleteClusterCmd returns a cobra.Command to delete a cluster from the license server
func DeleteClusterCmd() *cobra.Command {
	var id string
	deleteCluster := &cobra.Command{
		Short: "Delete a cluster registered with the license server.",
		Long:  "Delete a cluster registered with the license server.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}

			_, err = c.License.DeleteCluster(c.Ctx(), &license.DeleteClusterRequest{
				Id: id,
			})
			return grpcutil.ScrubGRPC(err)
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
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}

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
func DeleteAllCmd() *cobra.Command {
	activate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Delete all data from the license server",
		Long:  "Delete all data from the license server",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			if _, err := c.License.DeleteAll(c.Ctx(), &license.DeleteAllRequest{}); err != nil {
				return err
			}
			fmt.Printf("All data deleted from license server.")
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
		Run: cmdutil.Run(func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			resp, err := c.License.GetActivationCode(c.Ctx(), &license.GetActivationCodeRequest{})
			if err != nil {
				return err
			}
			if resp.State == enterprise.State_NONE {
				fmt.Println("No Pachyderm Enterprise license is configured")
				return nil
			}
			ts, err := types.TimestampFromProto(resp.Info.Expires)
			if err != nil {
				return errors.Wrapf(err, "expiration timestamp could not be parsed")
			}
			fmt.Printf("Pachyderm Enterprise token state: %s\nExpiration: %s\nLicense: %s\n",
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
