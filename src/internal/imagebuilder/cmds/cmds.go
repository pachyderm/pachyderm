package cmds

import (
	"fmt"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/imagebuilder/jobs"
	ibstar "github.com/pachyderm/pachyderm/v2/src/internal/imagebuilder/starlark"
	"github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "build",
	Short: "Manage builds.",
}

func fileMustExist(i int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if i >= len(args) {
			return nil
		}
		if _, err := os.Stat(args[i]); err != nil {
			return err
		}
		return nil
	}
}

var shell = &cobra.Command{
	Use:   "shell [<workflow.star>]",
	Short: "Load the named workflow, then drop into a debugging shell.",
	Args:  cobra.MatchAll(cobra.RangeArgs(0, 1), fileMustExist(0)),
	RunE: func(cmd *cobra.Command, args []string) error {
		var in string
		if len(args) == 1 {
			in = args[0]
		}
		return starlark.RunShell(cmd.Context(), in, starlark.Options{
			ThreadLocalVars: ibstar.ThreadLocals,
			REPLPredefined:  starlark.Union(ibstar.Module, ibstar.DebugModule),
		})
	},
}

var plan = &cobra.Command{
	Use:   "plan <workflow.star> ref [... refs]",
	Short: "Show the steps necessary to produce the referenced artifacts using the provided workflow.",
	Args:  cobra.MatchAll(cobra.MinimumNArgs(1), fileMustExist(0)),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := args[0]
		ctx := cmd.Context()
		if _, err := starlark.RunProgram(ctx, in, starlark.Options{
			ThreadLocalVars: ibstar.ThreadLocals,
		}); err != nil {
			return errors.Wrap(err, "eval")
		}
		var refs []jobs.Reference
		for i, x := range args[1:] {
			ref, err := jobs.ParseRef(x)
			if err != nil {
				return errors.Wrapf(err, "parse ref[%d]", i)
			}
			refs = append(refs, ref)
		}
		js := ibstar.ThreadLocals[jobs.StarlarkRegistryKey].(*jobs.Registry)
		plan, err := jobs.Plan(ctx, js.Jobs, refs)
		if err != nil {
			return errors.Wrap(err, "plan")
		}
		fmt.Println(plan.String())
		return nil
	},
}

var run = &cobra.Command{
	Use:   "run <workflow.star> ref [... refs]",
	Short: "Build the referenced artifacts using the provided workflow.",
	Args:  cobra.MatchAll(cobra.MinimumNArgs(1), fileMustExist(0)),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		in := args[0]
		ctx := cmd.Context()
		if _, err := starlark.RunProgram(ctx, in, starlark.Options{
			ThreadLocalVars: ibstar.ThreadLocals,
		}); err != nil {
			return errors.Wrap(err, "eval")
		}
		var refs []jobs.Reference
		for i, x := range args[1:] {
			ref, err := jobs.ParseRef(x)
			if err != nil {
				return errors.Wrapf(err, "parse ref[%d]", i)
			}
			refs = append(refs, ref)
		}
		js := ibstar.ThreadLocals[jobs.StarlarkRegistryKey].(*jobs.Registry)
		got, err := jobs.Resolve(ctx, js.Jobs, refs)
		if err != nil {
			return err
		}
		fmt.Println("Resulting artifacts:")
		for _, a := range got {
			fmt.Printf("    %v\n", a)
		}
		fmt.Printf("Build finished in %v\n", time.Since(start).Round(time.Millisecond))
		return nil
	},
}

func init() {
	Root.AddCommand(run, plan, shell)
}
