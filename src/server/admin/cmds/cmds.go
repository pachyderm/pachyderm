package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/golang/snappy"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing admin commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var noObjects bool
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
				return c.ExtractURL(url)
			}
			w := snappy.NewBufferedWriter(os.Stdout)
			defer func() {
				if err := w.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			return c.ExtractWriter(!noObjects, w)
		}),
	}
	extract.Flags().BoolVar(&noObjects, "no-objects", false, "don't extract from object storage, only extract data from etcd")
	extract.Flags().StringVarP(&url, "url", "u", "", "An object storage url (i.e. s3://...) to extract to.")
	commands = append(commands, cmdutil.CreateAlias(extract, "extract"))

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
