package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/cobra"
)

// ActivateCmd returns a cobra.Command to activate the enterprise features of
// Pachyderm within a Pachyderm cluster. All repos will go from
// publicly-accessible to accessible only by the owner, who can subsequently add
// users
func ActivateCmd() *cobra.Command {
	activate := &cobra.Command{
		Use: "activate activation-code",
		Short: "Activate the enterprise features of Pachyderm with an activation " +
			"code",
		Long: "Activate the enterprise features of Pachyderm with an activation " +
			"code",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			activationCode := args[0]
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %s", err.Error())
			}
			_, err = c.Enterprise.Activate(c.Ctx(),
				&enterprise.ActivateRequest{ActivationCode: activationCode})
			return err
		}),
	}
	return activate
}

// GetStateCmd returns a cobra.Command to activate the enterprise features of
// Pachyderm within a Pachyderm cluster. All repos will go from
// publicly-accessible to accessible only by the owner, who can subsequently add
// users
func GetStateCmd() *cobra.Command {
	getState := &cobra.Command{
		Use: "get-state",
		Short: "Check whether the Pachyderm cluster has enterprise features " +
			"activated",
		Long: "Check whether the Pachyderm cluster has enterprise features " +
			"activated",
		Run: cmdutil.Run(func(args []string) error {
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %s", err.Error())
			}
			resp, err := c.Enterprise.GetState(c.Ctx(), &enterprise.GetStateRequest{})
			if err != nil {
				return err
			}
			fmt.Println(resp.State.String())
			return nil
		}),
	}
	return getState
}

// Cmds returns pachctl commands related to Pachyderm Enterprise
func Cmds() []*cobra.Command {
	enterprise := &cobra.Command{
		Use:   "enterprise",
		Short: "Enterprise commands enable Pachyderm Enterprise features",
		Long:  "Enterprise commands enable Pachyderm Enterprise features",
	}
	enterprise.AddCommand(ActivateCmd())
	enterprise.AddCommand(GetStateCmd())
	return []*cobra.Command{enterprise}
}
