package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm"
	pfscmds "github.com/pachyderm/pachyderm/src/pfs/cmds"
	deploycmds "github.com/pachyderm/pachyderm/src/pkg/deploy/cmds"
	ppscmds "github.com/pachyderm/pachyderm/src/pps/cmds"
	"github.com/spf13/cobra"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/proto/version"
	"golang.org/x/net/context"
)

func PachctlCmd(pfsdAddress string, ppsdAddress string) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Long: `Access the Pachyderm API.

Envronment variables:
  PFS_ADDRESS=0.0.0.0:650, the PFS server to connect to.
  PPS_ADDRESS=0.0.0.0:651, the PPS server to connect to.
`,
	}
	pfsCmds := pfscmds.Cmds(pfsdAddress)
	for _, cmd := range pfsCmds {
		rootCmd.AddCommand(cmd)
	}
	ppsCmds, err := ppscmds.Cmds(ppsdAddress)
	if err != nil {
		return nil, err
	}
	for _, cmd := range ppsCmds {
		rootCmd.AddCommand(cmd)
	}

	deployCmds := deploycmds.Cmds()
	for _, cmd := range deployCmds {
		rootCmd.AddCommand(cmd)
	}
	version := &cobra.Command{
		Use:   "version",
		Short: "Return version information.",
		Long:  "Return version information.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			pfsdVersionClient, err := getVersionAPIClient(pfsdAddress)
			if err != nil {
				return err
			}
			pfsdVersion, err := pfsdVersionClient.GetVersion(context.Background(), &google_protobuf.Empty{})
			if err != nil {
				return err
			}
			ppsdVersionClient, err := getVersionAPIClient(pfsdAddress)
			if err != nil {
				return err
			}
			ppsdVersion, err := ppsdVersionClient.GetVersion(context.Background(), &google_protobuf.Empty{})
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			printVerisonHeader(writer)
			printVersion(writer, "pachctl", pachyderm.Version)
			printVersion(writer, "pfsd", pfsdVersion)
			printVersion(writer, "ppsd", ppsdVersion)
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
		"%s\t%d.%d.%d(%s)\t\n",
		component,
		version.Major,
		version.Minor,
		version.Micro,
		version.Additional,
	)
}
