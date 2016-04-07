package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/client"
	pfscmds "github.com/pachyderm/pachyderm/src/server/pfs/cmds"
	ppscmds "github.com/pachyderm/pachyderm/src/server/pps/cmds"
	"github.com/spf13/cobra"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/proto/version"
	"golang.org/x/net/context"
)

func PachctlCmd(address string) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Long: `Access the Pachyderm API.

Envronment variables:
  ADDRESS=0.0.0.0:30650, the server to connect to.
`,
	}
	pfsCmds := pfscmds.Cmds(address)
	for _, cmd := range pfsCmds {
		rootCmd.AddCommand(cmd)
	}
	ppsCmds, err := ppscmds.Cmds(address)
	if err != nil {
		return nil, err
	}
	for _, cmd := range ppsCmds {
		rootCmd.AddCommand(cmd)
	}

	version := &cobra.Command{
		Use:   "version",
		Short: "Return version information.",
		Long:  "Return version information.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			versionClient, err := getVersionAPIClient(address)
			if err != nil {
				return err
			}
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			version, err := versionClient.GetVersion(ctx, &google_protobuf.Empty{})
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			printVerisonHeader(writer)
			printVersion(writer, "pachctl", client.Version)
			printVersion(writer, "pachd", version)
			return writer.Flush()
		}),
	}
	rootCmd.AddCommand(version)
	return rootCmd, nil
}

func getVersionAPIClient(address string) (protoversion.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return protoversion.NewAPIClient(clientConn), nil
}

func printVerisonHeader(w io.Writer) {
	fmt.Fprintf(w, "COMPONENT\tVERSION\t\n")
}

func printVersion(w io.Writer, component string, version *protoversion.Version) {
	fmt.Fprintf(
		w,
		"%s\t%d.%d.%d",
		component,
		version.Major,
		version.Minor,
		version.Micro,
	)
	if version.Additional != "" {
		fmt.Fprintf(w, "(%s)", version.Additional)
	}
	fmt.Fprintf(w, "\t\n")
}
