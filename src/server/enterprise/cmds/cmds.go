package cmds

import (
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/license"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
)

// ActivateCmd returns a cobra.Command to activate the license service,
// register the current pachd and activate enterprise features.
// This only makes sense in a single-cluster enterprise environment where
// this is one pachd acting as the Enterprise Server.
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

			// Register the localhost as a cluster
			resp, err := c.License.AddCluster(c.Ctx(),
				&license.AddClusterRequest{
					Id:      "localhost",
					Address: "localhost:650",
				})
			if err != nil {
				return errors.Wrapf(err, "could not register pachd with the license service")
			}

			// activate the Enterprise service
			_, err = c.Enterprise.Activate(c.Ctx(),
				&enterprise.ActivateRequest{
					Id:            "localhost",
					Secret:        resp.Secret,
					LicenseServer: "localhost:650",
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
	var server, id string
	register := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Register the cluster with an enterprise license server",
		Long:  "Register the cluster with an enterprise license server",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			secret, err := cmdutil.ReadPassword("Secret: ")
			if err != nil {
				return errors.Wrapf(err, "could not read shared secret")
			}

			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			// activate the Enterprise service
			_, err = c.Enterprise.Activate(c.Ctx(),
				&enterprise.ActivateRequest{
					Id:            id,
					Secret:        strings.TrimSpace(secret),
					LicenseServer: server,
				})
			if err != nil {
				return errors.Wrapf(err, "could not register with the license server")
			}

			return nil
		}),
	}
	register.PersistentFlags().StringVar(&server, "server", "", "the Enterprise Server to register with")
	register.PersistentFlags().StringVar(&id, "id", "", "the id for this cluster")

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

	return commands
}
