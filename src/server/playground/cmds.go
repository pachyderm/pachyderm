package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"

	"github.com/spf13/cobra"
)

func LoadExample() *cobra.Command {
	var url string
	loadExample := &cobra.Command{
		Short: "Set the identity server config",
		Long:  `Set the identity server config`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			req := &identity.SetIdentityServerConfigRequest{
				Config: &identity.IdentityServerConfig{
					Issuer: issuer,
				}}

			_, err = c.SetIdentityServerConfig(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	setConfig.PersistentFlags().StringVar(&issuer, "issuer", "", `The issuer for the identity server.`)
	return cmdutil.CreateAlias(setConfig, "idp set-config")
}
