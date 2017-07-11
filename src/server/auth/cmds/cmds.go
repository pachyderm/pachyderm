package cmds

import (
	"context"

	pach "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Cmds returns auth commands.
func Cmds(address string, noMetrics *bool) ([]*cobra.Command, error) {
	metrics := !*noMetrics
	putToken := &cobra.Command{
		Use:   "put-token token signature",
		Short: "Activate enterprise features with a token.",
		Long:  "Activate enterprise features with a token.",
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			client, err := pach.NewMetricsClientFromAddress(address, metrics, "user")
			if err != nil {
				return err
			}
			if _, err := client.AuthAPIClient.PutActivationCode(context.Background(), &auth.ActivationCode{
				Token:     args[0],
				Signature: args[1],
			}); err != nil {
				return err
			}
			return nil
		}),
	}
	return []*cobra.Command{
		putToken,
	}, nil
}
