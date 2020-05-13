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
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/hanwen/go-fuse/v2/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fuse"
	"github.com/spf13/cobra"
)

func parseRepoOpts(args []string) (map[string]*fuse.RepoOptions, error) {
	result := make(map[string]*fuse.RepoOptions)
	for _, arg := range args {
		var repo string
		var flag string
		opts := &fuse.RepoOptions{}
		repoAndRest := strings.Split(arg, "@")
		if len(repoAndRest) == 1 {
			// No branch specified
			opts.Branch = "master"
			repoAndFlag := strings.Split(repoAndRest[0], "+")
			repo = repoAndFlag[0]
			if len(repoAndFlag) > 1 {
				flag = repoAndFlag[1]
			}
		} else {
			repo = repoAndRest[0]
			branchAndFlag := strings.Split(repoAndRest[1], "+")
			opts.Branch = branchAndFlag[0]
			if len(branchAndFlag) > 1 {
				flag = branchAndFlag[1]
			}
		}
		if flag != "" {
			for _, c := range flag {
				if c != 'w' && c != 'r' {
					return nil, errors.Errorf("invalid format %q: unrecognized mode: %q", arg, flag)
				}
			}
			if strings.Contains("w", flag) {
				opts.Write = true
			}
		}
		if repo == "" {
			return nil, errors.Errorf("invalid format %q: repo cannot be empty", arg)
		}
		result[repo] = opts
	}
	return result, nil
}

func mountCmds() []*cobra.Command {
	var commands []*cobra.Command

	var write bool
	var debug bool
	var repoOpts cmdutil.RepeatedStringArg
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
			repoOpts, err := parseRepoOpts(repoOpts)
			if err != nil {
				return err
			}
			opts := &fuse.Options{
				Write: write,
				Fuse: &fs.Options{
					MountOptions: gofuse.MountOptions{
						Debug: debug,
					},
				},
				RepoOptions: repoOpts,
			}
			return fuse.Mount(c, mountPoint, opts)
		}),
	}
	mount.Flags().BoolVarP(&write, "write", "w", false, "Allow writing to pfs through the mount.")
	mount.Flags().BoolVarP(&debug, "debug", "d", false, "Turn on debug messages.")
	mount.Flags().VarP(&repoOpts, "repos", "r", "Repos and branches / commits to mount, arguments should be of the form \"repo@branch+w\", where the trailing flag \"+w\" indicates write.")
	mount.MarkFlagCustom("repos", "__pachctl_get_repo_branch")
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
