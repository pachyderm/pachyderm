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

	var url string
	extract := &cobra.Command{
		Use:   "extract",
		Short: "Extract Pachyderm state to stdout.",
		Long:  "Extract Pachyderm state to stdout.",
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
			return c.ExtractWriter(w)
		}),
	}
	extract.Flags().StringVarP(&url, "url", "u", "", "An object storage url (i.e. s3://...) to extract to.")
	restore := &cobra.Command{
		Use:   "restore",
		Short: "Restore Pachyderm state from stdin.",
		Long:  "Restore Pachyderm state from stdin.",
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
