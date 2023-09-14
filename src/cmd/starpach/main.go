// Command starpach is a tool for developers of Pachyderm.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"sort"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"github.com/spf13/cobra"
	"go.starlark.net/starlark"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

var (
	verbose    bool
	profile    bool
	gofh, stfh *os.File
	logFile    string
	endLogging func(error)
	timeout    time.Duration
	root       = &cobra.Command{
		Use:           os.Args[0],
		Short:         "Invoke various tools that make working on Pachyderm easier.",
		SilenceUsage:  true, // This avoids usage on errors.
		SilenceErrors: true, // We print our own errors.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
			endLogging = log.InitBatchLogger(logFile)

			ctx, stopSignals := signal.NotifyContext(pctx.Background(""), os.Interrupt)
			stop := func() { stopSignals() }
			go func() {
				<-ctx.Done()
				stop()
			}()
			if timeout > 0 {
				var c func()
				ctx, c = context.WithTimeout(ctx, timeout)
				stop = func() { stopSignals(); c() }
			}
			cmd.SetContext(ctx)
			if profile {
				var err error
				gofh, err = os.Create(fmt.Sprintf("/tmp/pprof.%v.go.out", os.Getpid()))
				if err != nil {
					return errors.Wrap(err, "create file for cpu profile")
				}
				if err := pprof.StartCPUProfile(gofh); err != nil {
					log.Info(ctx, "failed to start go profiler", zap.Error(err))
				}
				stfh, err = os.Create(fmt.Sprintf("/tmp/pprof.%v.starlark.out", os.Getpid()))
				if err != nil {
					return errors.Wrap(err, "create file for cpu profile")
				}
				if err := starlark.StartProfile(stfh); err != nil {
					log.Info(ctx, "failed to start starlark profiler", zap.Error(err))
				}
			}
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) (retErr error) {
			if profile {
				defer errors.Close(&retErr, gofh, "close go profile")
				defer errors.Close(&retErr, stfh, "close starlark profile")
				pprof.StopCPUProfile()
				if err := starlark.StopProfile(); err != nil {
					return errors.Wrap(err, "stop starlark profiler")
				}
			}
			return nil
		},
	}

	run = &cobra.Command{
		Use:   "run <workflow.star>",
		Short: "Build the referenced artifacts using the provided workflow.",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cmdutil.FileMustExist(0)),
		RunE: func(cmd *cobra.Command, args []string) error {
			got, err := ourstar.RunProgram(cmd.Context(), args[0], ourstar.Personality.Options)
			if err != nil {
				return err
			}
			keys := maps.Keys(got)
			sort.Strings(keys)
			for _, k := range keys {
				v := got[k]
				if verbose {
					fmt.Fprintf(os.Stdout, "%v: %#v\n", k, v)
				} else {
					fmt.Fprintf(os.Stdout, "%v: %v\n", k, v)
				}
			}
			return nil
		},
	}

	shell = &cobra.Command{
		Use:   "shell [<workflow.star>]",
		Short: "Run the named Starlark program, then drop into a debugging shell.",
		Args:  cobra.MatchAll(cobra.RangeArgs(0, 1), cmdutil.FileMustExist(0)),
		RunE: func(cmd *cobra.Command, args []string) error {
			var in string
			if len(args) == 1 {
				in = args[0]
			}
			return ourstar.RunShell(cmd.Context(), in, ourstar.Personality.Options)
		},
	}
)

func init() {
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "If true, show all log messages.")
	root.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 0, "If non-zero, stop all work after this interval.")
	root.PersistentFlags().BoolVar(&profile, "profile", false, "If true, write a CPU profile to /tmp/pprof.*.out.")
	root.PersistentFlags().StringVar(&logFile, "log", "", "If set, also log to a file.")

	ourstar.UsePersonalityFlag(run.Flags())

	ourstar.UsePersonalityFlag(shell.Flags())

	root.AddCommand(run, shell)
}

func main() {
	err := root.ExecuteContext(pctx.Background(""))
	if err != nil {
		starErr := new(starlark.EvalError)
		if errors.As(err, &starErr) {
			fmt.Fprintf(os.Stderr, "%v\n", starErr.Backtrace())
			if goErr := starErr.Unwrap(); verbose && goErr != nil {
				v := fmt.Sprintf("%+v", goErr)
				if v != err.Error() {
					fmt.Fprintf(os.Stderr, "%v\n", v)
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
			if verbose {
				v := fmt.Sprintf("%+v", err)
				if v != err.Error() {
					fmt.Fprintf(os.Stderr, "%v\n", v)
				}
			}
		}
	}
	if endLogging != nil {
		endLogging(err)
	}
}
