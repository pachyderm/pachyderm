//go:build !windows
// +build !windows

package cmds

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fuse"
	"github.com/spf13/cobra"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/fuse"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
)

const (
	name = "pfs"
)

func parseRepoOpts(project string, args []string) (map[string]*fuse.RepoOptions, error) {
	result := make(map[string]*fuse.RepoOptions)
	for _, arg := range args {
		opts := &fuse.RepoOptions{}
		fileAndFlag := strings.Split(arg, "+")
		file, err := cmdutil.ParseFile(project, fileAndFlag[0])
		if err != nil {
			return nil, err
		}
		opts.Name = file.Commit.Branch.Repo.Name
		opts.File = file
		if len(fileAndFlag) > 1 {
			for _, c := range fileAndFlag[1] {
				if c != 'w' && c != 'r' {
					return nil, errors.Errorf("invalid format %q: unrecognized mode: %q", arg, c)
				}
			}
			if strings.Contains("w", fileAndFlag[1]) {
				opts.Write = true
			}
		}
		if opts.File.Commit.Branch.Name == "" {
			opts.File.Commit.Branch.Name = "master"
		}
		result[opts.File.Commit.Branch.Repo.Name] = opts
	}
	return result, nil
}

func mountCmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	var write bool
	var debug bool
	var repoOpts cmdutil.RepeatedStringArg
	var project string
	mount := &cobra.Command{
		Use:   "{{alias}} <path/to/mount/point>",
		Short: "Mount pfs locally. This command blocks.",
		Long:  "Mount pfs locally. This command blocks.",
		Run: cmdutil.RunFixedArgsCmd(1, func(cmd *cobra.Command, args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer c.Close()
			mountPoint := args[0]
			repoOpts, err := parseRepoOpts(project, repoOpts)
			if err != nil {
				return err
			}
			opts := &fuse.Options{
				Write: write,
				Fuse: &fs.Options{
					MountOptions: gofuse.MountOptions{
						Debug:  debug,
						FsName: name,
						Name:   name,
					},
				},
				RepoOptions: repoOpts,
			}
			// Prints a warning if we're on macOS
			PrintWarning()
			return fuse.Mount(c, project, mountPoint, opts)
		}),
	}
	mount.Flags().BoolVarP(&write, "write", "w", false, "Allow writing to pfs through the mount.")
	mount.Flags().BoolVarP(&debug, "debug", "d", false, "Turn on debug messages.")
	mount.Flags().VarP(&repoOpts, "repos", "r", "Repos and branches / commits to mount, arguments should be of the form \"repo[@branch=commit][+w]\", where the trailing flag \"+w\" indicates write. You can omit the branch when specifying a commit unless the same commit ID is on multiple branches in the repo.")
	mount.MarkFlagCustom("repos", "__pachctl_get_repo_branch")
	mount.Flags().StringVar(&project, "project", pfs.DefaultProjectName, "Project in which repo is located.")
	commands = append(commands, cmdutil.CreateAlias(mount, "mount"))

	var all bool
	unmount := &cobra.Command{
		Use:   "{{alias}} <path/to/mount/point>",
		Short: "Unmount pfs.",
		Long:  "Unmount pfs.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			if len(args) == 1 {
				return errors.EnsureStack(syscall.Unmount(args[0], 0))
			}
			if all {
				stdin := strings.NewReader(fmt.Sprintf(`
					mount | grep -w %s | grep fuse | cut -f 3 -d " "`,
					name,
				))
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
				fmt.Printf("This will unmount the following filesystems:\n")
				for _, mount := range mounts {
					fmt.Printf("%s\n", mount)
				}

				if ok, err := cmdutil.InteractiveConfirm(); err != nil {
					return err
				} else if !ok {
					return errors.New("unmount aborted")
				}

				for _, mount := range mounts {
					if err := syscall.Unmount(mount, 0); err != nil {
						return errors.EnsureStack(err)
					}
				}
			} else {
				return errors.Errorf("nothing to unmount specify a mounted filesystem or --all")
			}
			return nil
		}),
	}
	unmount.Flags().BoolVarP(&all, "all", "a", false, "unmount all pfs mounts")
	commands = append(commands, cmdutil.CreateAlias(unmount, "unmount"))

	return commands
}
