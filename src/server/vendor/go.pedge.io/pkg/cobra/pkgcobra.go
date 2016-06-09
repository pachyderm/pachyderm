package pkgcobra //import "go.pedge.io/pkg/cobra"

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Bounds represent min and max bounds.
type Bounds struct {
	Min int
	Max int
}

// RunFixedArgs makes a new cobra run function that checks that the number of args is equal to numArgs.
func RunFixedArgs(numArgs int, run func(args []string) error) func(*cobra.Command, []string) {
	return RunBoundedArgs(Bounds{Min: numArgs, Max: numArgs}, run)
}

// RunBoundedArgs makes a new cobra run function that checks that the number of args is within argBounds.
func RunBoundedArgs(argBounds Bounds, run func(args []string) error) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, args []string) {
		Check(CheckBoundedArgs(argBounds, args))
		Check(run(args))
	}
}

// Run makes a new cobra run function that wraps the given function.
func Run(run func(args []string) error) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, args []string) {
		Check(run(args))
	}
}

// CheckFixedArgs checks that the number of arguments equals numArgs.
func CheckFixedArgs(numArgs int, args []string) error {
	return CheckBoundedArgs(Bounds{Min: numArgs, Max: numArgs}, args)
}

// CheckBoundedArgs checks that the number of arguments is within the given Bounds.
func CheckBoundedArgs(argBounds Bounds, args []string) error {
	if argBounds.Min == 0 && argBounds.Max == 0 {
		if len(args) != 0 {
			return fmt.Errorf("Expected no args, got %d", len(args))
		}
		return nil
	}
	if argBounds.Max == 0 {
		if len(args) < argBounds.Min {
			return fmt.Errorf("Expected at least %d args, got %d", argBounds.Min, len(args))
		}
		return nil
	}
	if argBounds.Min == 0 {
		if len(args) > argBounds.Max {
			return fmt.Errorf("Expected at most %d args, got %d", argBounds.Max, len(args))
		}
		return nil
	}
	if len(args) < argBounds.Min || len(args) > argBounds.Max {
		if argBounds.Min == argBounds.Max {
			return fmt.Errorf("Expected %d args, got %d", argBounds.Min, len(args))
		}
		return fmt.Errorf("Expected between %d and %d args, got %d", argBounds.Min, argBounds.Max, len(args))
	}
	return nil
}

// Check checks the error.
func Check(err error) {
	if err != nil {
		ErrorAndExit(err.Error())
	}
}

// ErrorAndExit errors with the given format and args, and then exits.
func ErrorAndExit(format string, args ...interface{}) {
	if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
		fmt.Fprintf(os.Stderr, "%s\n", errString)
	}
	os.Exit(1)
}
