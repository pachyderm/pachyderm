// Command starpach is a tool for developers of Pachyderm.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	buildcmds "github.com/pachyderm/pachyderm/v2/src/internal/imagebuilder/cmds"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/spf13/cobra"
	"go.starlark.net/starlark"
)

var (
	verbose    bool
	profile    bool
	profileF   *os.File
	logFile    string
	endLogging func(error)
	timeout    time.Duration
	root       = &cobra.Command{
		Use:           os.Args[0],
		Short:         "Invoke various developer utilities.",
		SilenceUsage:  true, // This avoids usage on errors.
		SilenceErrors: true, // We print our own errors.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
			endLogging = log.InitBatchLogger(logFile)
			if timeout > 0 {
				ctx, _ := context.WithTimeout(cmd.Context(), timeout)
				cmd.SetContext(ctx)
			}
			if profile {
				var err error
				profileF, err = os.Create("/tmp/pprof.go.out")
				if err != nil {
					return errors.Wrap(err, "create file for cpu profile")
				}
				if err := pprof.StartCPUProfile(profileF); err != nil {
					return errors.Wrap(err, "start cpu profile")
				}
			}
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if profile {
				pprof.StopCPUProfile()
				if err := profileF.Close(); err != nil {
					return errors.Wrap(err, "close cpu profile file")
				}
			}
			return nil
		},
	}
)

func init() {
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "If true, show all log messages.")
	root.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 0, "If non-zero, stop all work after this interval.")
	root.PersistentFlags().BoolVar(&profile, "profile", false, "If true, write a CPU profile to /tmp/pprof.*.out.")
	root.PersistentFlags().StringVar(&logFile, "log", "", "If set, also log to a file.")
	root.AddCommand(buildcmds.Root)
}

func main() {
	// TODO: always log to a file; delete file if command exits 0.
	ctx, stop := signal.NotifyContext(pctx.Background(""), os.Interrupt)
	go func() {
		<-ctx.Done()
		stop()
	}()
	err := root.ExecuteContext(ctx)
	if err != nil {
		starErr := new(starlark.EvalError)
		if errors.As(err, &starErr) {
			fmt.Fprintf(os.Stderr, "%v\n", starErr.Backtrace())
		} else {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
		if verbose {
			v := fmt.Sprintf("%+v", err)
			if v != err.Error() {
				fmt.Fprintf(os.Stderr, "%v\n", v)
			}
		}
	}
	if endLogging != nil {
		endLogging(err)
	}
}
