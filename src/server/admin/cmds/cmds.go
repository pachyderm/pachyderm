package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"

	"github.com/golang/snappy"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing admin commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var noObjects, noEnterprise, noAuth bool
	var url string
	extract := &cobra.Command{
		Short: "Extract Pachyderm state to stdout or an object store bucket.",
		Long:  "Extract Pachyderm state to stdout or an object store bucket.",
		Example: `
# Extract into a local file:
$ {{alias}} > backup

# Extract to s3:
$ {{alias}} -u s3://bucket/backup`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}

			defer c.Close()
			if url != "" {
				return c.ExtractURL(url, !noEnterprise, !noAuth)
			}
			w := snappy.NewBufferedWriter(os.Stdout)
			defer func() {
				if err := w.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			return c.ExtractWriter(!noObjects, !noEnterprise, !noAuth, w)
		}),
	}
	extract.Flags().BoolVar(&noObjects, "no-objects", false, "don't extract from object storage, only extract data from etcd")
	extract.Flags().BoolVar(&noEnterprise, "no-enterprise", false, "don't extract the enterprise license information")
	extract.Flags().BoolVar(&noAuth, "no-auth", false, "don't extract the authentication information")
	extract.Flags().StringVarP(&url, "url", "u", "", "An object storage url (i.e. s3://...) to extract to.")
	commands = append(commands, cmdutil.CreateAlias(extract, "extract"))

	var noToken bool
	restore := &cobra.Command{
		Short: "Restore Pachyderm state from stdin or an object store.",
		Long:  "Restore Pachyderm state from stdin or an object store.",
		Example: `
# Restore from a local file:
$ {{alias}} < backup

# Restore from s3:
$ {{alias}} -u s3://bucket/backup`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}

			// Generate a root auth token and cache it in the current context. If the restore operation
			// activates auth this will become the new root token.
			if !noToken {
				authToken := uuid.NewWithoutDashes()
				if err := config.WritePachTokenToConfig(authToken); err != nil {
					return err
				}
				c.SetAuthToken(authToken)

				defer func() {
					authActive, err := c.IsAuthActive()
					if err != nil {
						fmt.Fprintf(os.Stderr, "Unable to detect whether auth is active on the restored cluster - %v", err)
					}

					if authActive || err != nil {
						fmt.Fprintf(os.Stderr, `
Generated a new root auth token. This token has permanent admin access to the restored cluster.
It cannot be recreated - keep it safe for the lifetime of the cluster.

Token: %v
`[1:], authToken)
					}
				}()
			}

			defer c.Close()
			if url != "" {
				err = c.RestoreURL(url)
			} else {
				err = c.RestoreReader(snappy.NewReader(os.Stdin))
			}
			if err != nil {
				return errors.Wrapf(err, "WARNING: Your cluster might be in an invalid "+
					"state--consider deleting partially-restored data before continuing")
			}

			return nil
		}),
	}
	restore.Flags().StringVarP(&url, "url", "u", "", "An object storage url (i.e. s3://...) to restore from.")
	restore.Flags().BoolVar(&noToken, "no-token", false, "Don't generate a new auth token at the beginning of the restore.")
	commands = append(commands, cmdutil.CreateAlias(restore, "restore"))

	inspectCluster := &cobra.Command{
		Short: "Returns info about the pachyderm cluster",
		Long:  "Returns info about the pachyderm cluster",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()
			ci, err := c.InspectCluster()
			if err != nil {
				return err
			}
			fmt.Println(ci.ID)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(inspectCluster, "inspect cluster"))

	return commands
}
