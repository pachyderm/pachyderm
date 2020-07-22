package cmds

import (
	"os"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing debug commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var duration time.Duration
	var pachd bool
	var pipeline string
	var worker string
	profile := &cobra.Command{
		Use:   "{{alias}} <profile>",
		Short: "Collect a set of pprof profiles.",
		Long:  "Collect a set of pprof profiles.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-profile")
			if err != nil {
				return err
			}
			defer client.Close()
			var d *types.Duration
			if duration != 0 {
				d = types.DurationProto(duration)
			}
			p := &debug.Profile{
				Name:     args[0],
				Duration: d,
			}
			f, err := createFilter(pachd, pipeline, worker)
			if err != nil {
				return err
			}
			return client.Profile(p, f, os.Stdout)
		}),
	}
	profile.Flags().DurationVarP(&duration, "duration", "d", time.Minute, "Duration to run a CPU profile for.")
	profile.Flags().BoolVar(&pachd, "pachd", false, "Only collect the profile from pachd.")
	profile.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Only collect the profile from the worker pods for the given pipeline. (Note: Use the replication controller name.)")
	profile.Flags().StringVarP(&worker, "worker", "w", "", "Only collect the profile from the given worker pod.")
	commands = append(commands, cmdutil.CreateAlias(profile, "debug profile"))

	binary := &cobra.Command{
		Short: "Collect a set of binaries.",
		Long:  "Collect a set of binaries.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-binary")
			if err != nil {
				return err
			}
			defer client.Close()
			f, err := createFilter(pachd, pipeline, worker)
			if err != nil {
				return err
			}
			return client.Binary(f, os.Stdout)
		}),
	}
	binary.Flags().BoolVar(&pachd, "pachd", false, "Only collect the binary from pachd.")
	binary.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Only collect the binary from the worker pods for the given pipeline. (Note: Use the replication controller name.)")
	binary.Flags().StringVarP(&worker, "worker", "w", "", "Only collect the binary from the given worker pod.")
	commands = append(commands, cmdutil.CreateAlias(binary, "debug binary"))

	dump := &cobra.Command{
		Short: "Collect a standard set of debugging information.",
		Long:  "Collect a standard set of debugging information.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine("debug-dump")
			if err != nil {
				return err
			}
			defer client.Close()
			f, err := createFilter(pachd, pipeline, worker)
			if err != nil {
				return err
			}
			return client.Dump(f, os.Stdout)
		}),
	}
	dump.Flags().BoolVar(&pachd, "pachd", false, "Only collect the dump from pachd.")
	dump.Flags().StringVarP(&pipeline, "pipeline", "p", "", "Only collect the dump from the worker pods for the given pipeline. (Note: Use the replication controller name.)")
	dump.Flags().StringVarP(&worker, "worker", "w", "", "Only collect the dump from the given worker pod.")
	commands = append(commands, cmdutil.CreateAlias(dump, "debug dump"))

	debug := &cobra.Command{
		Short: "Debug commands for analyzing a running cluster.",
		Long:  "Debug commands for analyzing a running cluster.",
	}
	commands = append(commands, cmdutil.CreateAlias(debug, "debug"))

	return commands
}

func createFilter(pachd bool, pipeline, worker string) (*debug.Filter, error) {
	var f *debug.Filter
	if pachd {
		f = &debug.Filter{Filter: &debug.Filter_Pachd{Pachd: true}}
	}
	if pipeline != "" {
		if f != nil {
			return nil, errors.Errorf("only one debug filter allowed")
		}
		f = &debug.Filter{Filter: &debug.Filter_Pipeline{&debug.Pipeline{Name: pipeline}}}
	}
	if worker != "" {
		if f != nil {
			return nil, errors.Errorf("only one debug filter allowed")
		}
		f = &debug.Filter{Filter: &debug.Filter_Worker{&debug.Worker{Pod: worker}}}
	}
	return f, nil
}
