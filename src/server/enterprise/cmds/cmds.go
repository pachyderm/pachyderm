package cmds

import (
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
)

// Unfortunately, Go's pre-defined format strings for parsing RFC-3339-compliant
// timestamps aren't exhaustive. This method attempts to parse a larger set of
// of ISO-8601-compatible timestampts (which are themselves a subset of RFC-3339
// timestamps)
func parseISO8601(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("2006-01-02T15:04:05Z0700", s)
	if err == nil {
		return t, nil
	}
	return time.Time{}, errors.Errorf("could not parse \"%s\" as any of %v", s, []string{time.RFC3339, "2006-01-02T15:04:05Z0700"})
}

// ActivateCmd returns a cobra.Command to activate the enterprise features of
// Pachyderm within a Pachyderm cluster. All repos will go from
// publicly-accessible to accessible only by the owner, who can subsequently add
// users
func ActivateCmd() *cobra.Command {
	var expires string
	activate := &cobra.Command{
		Use: "{{alias}} <activation-code>",
		Short: "Activate the enterprise features of Pachyderm with an activation " +
			"code",
		Long: "Activate the enterprise features of Pachyderm with an activation " +
			"code",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			// request the enterprise key
			key, err := cmdutil.ReadPassword("Enterprise key: ")
			if err != nil {
				return errors.Wrapf(err, "could not read enterprise key")
			}

			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			req := &enterprise.ActivateRequest{}
			req.ActivationCode = key
			if expires != "" {
				t, err := parseISO8601(expires)
				if err != nil {
					return errors.Wrapf(err, "could not parse the timestamp \"%s\"", expires)
				}
				req.Expires, err = types.TimestampProto(t)
				if err != nil {
					return errors.Wrapf(err, "error converting expiration time \"%s\"", t.String())
				}
			}
			resp, err := c.Enterprise.Activate(c.Ctx(), req)
			if err != nil {
				return err
			}
			ts, err := types.TimestampFromProto(resp.Info.Expires)
			if err != nil {
				return errors.Wrapf(err, "activation request succeeded, but could not "+
					"convert token expiration time to a timestamp", err)
			}
			fmt.Printf("Activation succeeded. Your Pachyderm Enterprise token "+
				"expires %s\n", ts.String())
			return nil
		}),
	}
	activate.PersistentFlags().StringVar(&expires, "expires", "", "A timestamp "+
		"indicating when the token provided above should expire (formatted as an "+
		"RFC 3339/ISO 8601 datetime). This is only applied if it's earlier than "+
		"the signed expiration time encoded in 'activation-code', and therefore "+
		"is only useful for testing.")

	return cmdutil.CreateAlias(activate, "enterprise activate")
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
	commands = append(commands, GetStateCmd())

	return commands
}
