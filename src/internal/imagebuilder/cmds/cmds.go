package cmds

import (
	"fmt"
	"os"

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
	Use:   "plan <workflow.star>",
	Short: "Show the steps necessary to run the build specified by the named starpach build file.",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), fileMustExist(0)),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := args[0]
		ctx := cmd.Context()
		out, err := starlark.RunProgram(ctx, in, starlark.Options{
			ThreadLocalVars: ibstar.ThreadLocals,
		})
		if err != nil {
			return err
		}
		fmt.Printf("%v", out)
		return nil
	},
}

func init() {
	Root.AddCommand(plan, shell)
}
