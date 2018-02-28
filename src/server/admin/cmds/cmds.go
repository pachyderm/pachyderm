package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/golang/snappy"
	"github.com/spf13/cobra"
)

const (
	codestart = "```sh\n\n"
	codeend   = "\n```"
)

// Cmds returns a slice containing admin commands.
func Cmds(noMetrics *bool) []*cobra.Command {
	metrics := !*noMetrics

	var noObjects bool
	var url string
	extract := &cobra.Command{
		Use:   "extract",
		Short: "Extract Pachyderm state to stdout or an object store bucket.",
		Long: `Extract Pachyderm state to stdout or an object store bucket.
` + codestart + `# Extract into a local file:
pachctl extract >backup

# Extract to s3:
pachctl extract -u s3://bucket/backup` + codeend,
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
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
	restore := &cobra.Command{
		Use:   "restore",
		Short: "Restore Pachyderm state from stdin or an object store.",
		Long: `Restore Pachyderm state from stdin or an object store..
` + codestart + `# Restore from a local file:
pachctl restore <backup

# Restore from s3:
pachctl restore -u s3://bucket/backup` + codeend,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			if url != "" {
				return c.RestoreURL(url)
			}
			return c.RestoreReader(snappy.NewReader(os.Stdin))
		}),
	}
	restore.Flags().StringVarP(&url, "url", "u", "", "An object storage url (i.e. s3://...) to restore from.")
	inspectCluster := &cobra.Command{
		Use:   "inspect-cluster",
		Short: "Returns info about the pachyderm cluster",
		Long:  "Returns info about the pachyderm cluster",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			ci, err := c.InspectCluster()
			if err != nil {
				return err
			}
			fmt.Println(ci.ID)
			return nil
		}),
	}
	return []*cobra.Command{extract, restore, inspectCluster}
}
