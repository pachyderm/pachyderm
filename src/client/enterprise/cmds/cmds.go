package cmds

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/cobra"
)

// ActivateCmd returns a cobra.Command to activate the enterprise features of
// Pachyderm within a Pachyderm cluster. All repos will go from
// publicly-accessible to accessible only by the owner, who can subsequently add
// users
func ActivateCmd() *cobra.Command {
	activate := &cobra.Command{
		Use: "activate-pachyderm-enterprise activation-code",
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
			// _, err = c.AuthAPIClient.ActivateEnterprise(c.Ctx(),
			// 	&auth.ActivateEnterpriseRequest{
			// 		ActivationCode: activationCode,
			// 	})
			// return err
		}),
	}
	return activate
}

// Cmds returns pachctl commands related to Pachyderm Enterprise
func Cmds() []*cobra.Command {
	enterprise := &cobra.Command{
		Use:   "enterprise",
		Short: "Enterprise commands enable Pachyderm Enterprise features",
		Long:  "Enterprise commands enable Pachyderm Enterprise features",
	}
	enterprise.AddCommand(ActivateCmd())
	return []*cobra.Command{enterprise}
}
