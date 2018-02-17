package cmds

import (
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/golang/snappy"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing admin commands.
func Cmds(noMetrics *bool) []*cobra.Command {
	metrics := !*noMetrics

	var noObjects bool
	var url string
	extract := &cobra.Command{
		Use:   "extract",
		Short: "Extract Pachyderm state to stdout or an s3 bucket.",
		Long:  "Extract Pachyderm state to stdout or an s3 bucket.",
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
		Short: "Restore Pachyderm state from stdin or an s3 bucket.",
		Long:  "Restore Pachyderm state from stdin or an s3 bucket.",
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
	return []*cobra.Command{extract, restore}
}
