package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/version"
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
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			printVersionHeader(writer)
			printVersion(writer, "pachctl", version.Version)
			writer.Flush()

			versionClient, err := getVersionAPIClient(address)
			if err != nil {
				return err
			}
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			version, err := versionClient.GetVersion(ctx, &google_protobuf.Empty{})

			if err != nil {
				fmt.Fprintf(writer, "pachd\tUNKNOWN: Error %v\n", err)
				return writer.Flush()
			}

			printVersion(writer, "pachd", version)
			return writer.Flush()
		}),
	}
	deleteAll := &cobra.Command{
		Use:   "delete-all",
		Short: "Delete everything.",
		Long:  "Delete everything.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return err
			}
			fmt.Printf("Are you sure you want to delete everything? yN\n")
			r := bufio.NewReader(os.Stdin)
			bytes, err := r.ReadBytes('\n')
			if err != nil {
				return err
			}
			if bytes[0] == 'y' || bytes[0] == 'Y' {
				return client.DeleteAll()
			}
			return nil
		}),
	}
	rootCmd.AddCommand(version)
	rootCmd.AddCommand(deleteAll)
	return rootCmd, nil
}

func getVersionAPIClient(address string) (protoversion.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return protoversion.NewAPIClient(clientConn), nil
}

func printVersionHeader(w io.Writer) {
	fmt.Fprintf(w, "COMPONENT\tVERSION\t\n")
}

func printVersion(w io.Writer, component string, v *protoversion.Version) {
	fmt.Fprintf(w, "%s\t%s\t\n", component, version.PrettyPrintVersion(v))
}
