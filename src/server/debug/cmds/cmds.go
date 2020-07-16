package cmds

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing debug commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	dump := &cobra.Command{
		Short: "Return a dump of running goroutines.",
		Long:  "Return a dump of running goroutines.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.Dump(os.Stdout)
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(dump, "debug dump"))

	var duration time.Duration
	var workerStr string
	profile := &cobra.Command{
		Use:   "{{alias}} <profile>",
		Short: "Return a profile from the server.",
		Long:  "Return a profile from the server.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			var worker *debug.Worker
			if workerStr != "" {
				var err error
				worker, err = parseWorker(workerStr)
				if err != nil {
					return err
				}
			}
			client, err := client.NewOnUserMachine("debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.Profile(args[0], duration, worker, os.Stdout)
		}),
	}
	profile.Flags().DurationVarP(&duration, "duration", "d", time.Minute, "Duration to run a CPU profile for.")
	profile.Flags().StringVarP(&workerStr, "worker", "w", "", "Collect the profile from the given worker pod and container. (format: <pod>:<container>)")
	commands = append(commands, cmdutil.CreateAlias(profile, "debug profile"))

	binary := &cobra.Command{
		Short: "Return the binary the server is running.",
		Long:  "Return the binary the server is running.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			return client.Binary(os.Stdout)
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(binary, "debug binary"))

	var profileFile string
	var binaryFile string
	var interactive bool
	pprof := &cobra.Command{
		Use:   "{{alias}} <profile>",
		Short: "Analyze a profile of pachd in pprof.",
		Long:  "Analyze a profile of pachd in pprof.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			var eg errgroup.Group
			// Download the profile
			eg.Go(func() (retErr error) {
				if args[0] == "cpu" {
					fmt.Printf("Downloading cpu profile, this will take %s...", units.HumanDuration(duration))
				}
				f, err := os.Create(profileFile)
				if err != nil {
					return err
				}
				defer func() {
					if err := f.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				return client.Profile(args[0], duration, nil, f)
			})
			// Download the binary
			eg.Go(func() (retErr error) {
				f, err := os.Create(binaryFile)
				if err != nil {
					return err
				}
				defer func() {
					if err := f.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				return client.Binary(f)
			})
			if err := eg.Wait(); err != nil {
				return err
			}
			if interactive {
				cmd := exec.Command("go", "tool", "pprof", binaryFile, profileFile)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return cmd.Run()
			}
			return nil
		}),
	}
	pprof.Flags().StringVar(&profileFile, "profile-file", "profile", "File to write the profile to.")
	pprof.Flags().StringVar(&binaryFile, "binary-file", "binary", "File to write the binary to.")
	pprof.Flags().DurationVarP(&duration, "duration", "d", time.Minute, "Duration to run a CPU profile for.")
	pprof.Flags().BoolVarP(&interactive, "interactive", "i", false, "If set, "+
		"open an interactive session with 'go tool pprof' analyzing the collected "+
		"profile.")
	commands = append(commands, cmdutil.CreateAlias(pprof, "debug pprof"))

	debug := &cobra.Command{
		Short: "Debug commands for analyzing a running cluster.",
		Long:  "Debug commands for analyzing a running cluster.",
	}
	commands = append(commands, cmdutil.CreateAlias(debug, "debug"))

	return commands
}

func parseWorker(worker string) (*debug.Worker, error) {
	strs := strings.Split(worker, ":")
	if len(strs) != 2 {
		return nil, errors.Errorf("error parsing worker string: %v (format: <pod>:<container>)", worker)
	}
	return &debug.Worker{
		Pod:       strs[0],
		Container: strs[1],
	}, nil
}
