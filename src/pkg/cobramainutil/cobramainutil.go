package cobramainutil

import (
	"fmt"
	"math"
	"os"

	"github.com/pachyderm/pachyderm/src/pkg/protoversion"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type Command struct {
	Use        string
	Long       string
	NumArgs    int
	MinNumArgs int
	MaxNumArgs int
	Run        func(*cobra.Command, []string) error
}

func (c Command) ToCobraCommand() *cobra.Command {
	return &cobra.Command{
		Use:  c.Use,
		Long: c.Long,
		Run: func(cmd *cobra.Command, args []string) {
			if c.NumArgs > 0 {
				if c.MinNumArgs > 0 || c.MaxNumArgs > 0 {
					check(fmt.Errorf("system error: cannot specify NumArgs and MinNumArgs or MaxNumArgs"))
				}
				if err := checkArgs(args, c.NumArgs, c.Use); err != nil {
					check(err)
				}
			} else {
				minExpected := c.MinNumArgs
				maxExpected := c.MaxNumArgs
				if maxExpected == 0 {
					maxExpected = math.MaxInt32
				}
				if err := checkArgsBounds(args, minExpected, maxExpected, c.Use); err != nil {
					check(err)
				}
			}
			check(c.Run(cmd, args))
		},
	}
}

func NewVersionCommand(clientConn *grpc.ClientConn, clientVersion *protoversion.Version) *cobra.Command {
	return Command{
		Use:  "version",
		Long: "Print the version.",
		Run: func(cmd *cobra.Command, args []string) error {
			serverVersion, err := protoversion.GetVersion(protoversion.NewApiClient(clientConn))
			if err != nil {
				return err
			}
			fmt.Printf("Client: %s\nServer: %s\n", clientVersion.VersionString(), serverVersion.VersionString())
			return nil
		},
	}.ToCobraCommand()
}

func checkArgs(args []string, expected int, usage string) error {
	if len(args) != expected {
		return fmt.Errorf("Wrong number of arguments. Got %d, need %d. %s\n", len(args), expected, usage)
	}
	return nil
}

func checkArgsBounds(args []string, minExpected int, maxExpected int, usage string) error {
	if len(args) < minExpected || len(args) > maxExpected {
		return fmt.Errorf("Wrong number of arguments. Got %d, need %d to %d. %s\n", len(args), minExpected, maxExpected, usage)
	}
	return nil
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
