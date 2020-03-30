// +build !windows

package cmds

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse2"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/hanwen/go-fuse/v2/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fuse"
	"github.com/spf13/cobra"
)

func parseCommits(args []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, arg := range args {
		split := strings.Split(arg, "@")
		if len(split) != 2 {
			return nil, errors.Errorf("malformed input %s, must be of the form repo@commit", args)
		}
		result[split[0]] = split[1]
	}
	return result, nil
}

func mountCmds() []*cobra.Command {
	var commands []*cobra.Command

	var debug bool
	var commits cmdutil.RepeatedStringArg
	mount := &cobra.Command{
		Use:   "{{alias}} <path/to/mount/point>",
		Short: "Mount pfs locally. This command blocks.",
		Long:  "Mount pfs locally. This command blocks.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine("fuse")
			if err != nil {
				return err
			}
			defer c.Close()
			mountPoint := args[0]
			commits, err := parseCommits(commits)
			if err != nil {
				return err
			}
			opts := &fuse.Options{
				Fuse: &fs.Options{
					MountOptions: gofuse.MountOptions{
						Debug: debug,
					},
				},
				Commits: commits,
			}
			return fuse.Mount(c, mountPoint, opts)
		}),
	}
	mount.Flags().BoolVarP(&debug, "debug", "d", false, "Turn on debug messages.")
	mount.Flags().VarP(&commits, "commits", "c", "Commits to mount for repos, arguments should be of the form \"repo@commit\"")
	mount.MarkFlagCustom("commits", "__pachctl_get_repo_branch")
	commands = append(commands, cmdutil.CreateAlias(mount, "mount"))

	var all bool
	unmount := &cobra.Command{
		Use:   "{{alias}} <path/to/mount/point>",
		Short: "Unmount pfs.",
		Long:  "Unmount pfs.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			if len(args) == 1 {
				return syscall.Unmount(args[0], 0)
			}

			if all {
				stdin := strings.NewReader(`
		mount | grep pfs:// | cut -f 3 -d " "
		`)
				var stdout bytes.Buffer
				if err := cmdutil.RunIO(cmdutil.IO{
					Stdin:  stdin,
					Stdout: &stdout,
					Stderr: os.Stderr,
				}, "sh"); err != nil {
					return err
				}
				scanner := bufio.NewScanner(&stdout)
				var mounts []string
				for scanner.Scan() {
					mounts = append(mounts, scanner.Text())
				}
				if len(mounts) == 0 {
					fmt.Println("No mounts found.")
					return nil
				}
				fmt.Printf("Unmount the following filesystems? yN\n")
				for _, mount := range mounts {
					fmt.Printf("%s\n", mount)
				}
				r := bufio.NewReader(os.Stdin)
				bytes, err := r.ReadBytes('\n')
				if err != nil {
					return err
				}
				if bytes[0] == 'y' || bytes[0] == 'Y' {
					for _, mount := range mounts {
						if err := syscall.Unmount(mount, 0); err != nil {
							return err
						}
					}
				}
			}
			return nil
		}),
	}
	unmount.Flags().BoolVarP(&all, "all", "a", false, "unmount all pfs mounts")
	commands = append(commands, cmdutil.CreateAlias(unmount, "unmount"))

	return commands
}
