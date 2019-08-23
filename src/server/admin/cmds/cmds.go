package cmds

import (
	"fmt"
	"io"
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
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
				return fmt.Errorf("%v\nWARNING: Your cluster might be in an invalid "+
					"state--consider deleting partially-restored data before continuing",
					err)
			}
			return nil
		}),
	}
	restore.Flags().StringVarP(&url, "url", "u", "", "An object storage url (i.e. s3://...) to restore from.")
	commands = append(commands, cmdutil.CreateAlias(restore, "restore"))

	view := &cobra.Command{
		Short:   "View an extract.",
		Long:    "View an extract.",
		Example: "",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			reader := pbutil.NewReader(snappy.NewReader(os.Stdin))
			op := &admin.Op{}
			for {
				if err := reader.Read(op); err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				fmt.Println(op)
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(view, "view"))

	fix := &cobra.Command{
		Short:   "Fix an extract.",
		Long:    "Fix an extract.",
		Example: "",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			reader := pbutil.NewReader(snappy.NewReader(os.Stdin))
			out := pbutil.NewWriter(snappy.NewWriter(os.Stdout))
			op := &admin.Op{}
			pipelines := make(map[string]bool)
			for {
				if err := reader.Read(op); err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
				if op.Op1_8.Pipeline != nil {
					pipelines[op.Op1_8.Pipeline.Pipeline.Name] = true
				}
				if op.Op1_8.Commit != nil {
					if pipelines[op.Op1_8.Commit.Parent.Repo.Name] {
						continue
					}
				}
				if op.Op1_8.Branch != nil {
					if pipelines[op.Op1_8.Branch.Branch.Repo.Name] {
						continue
					}
				}
				if _, err := out.Write(op); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(fix, "fix"))

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
