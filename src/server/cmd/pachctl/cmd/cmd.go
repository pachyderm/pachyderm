package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/version"
	pfscmds "github.com/pachyderm/pachyderm/src/server/pfs/cmds"
	deploycmds "github.com/pachyderm/pachyderm/src/server/pkg/deploy/cmds"
	ppscmds "github.com/pachyderm/pachyderm/src/server/pps/cmds"
	"github.com/spf13/cobra"
	"go.pedge.io/lion"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/pkg/exec"
	"go.pedge.io/proto/version"
	"golang.org/x/net/context"
)

// PachctlCmd takes a pachd host-address and creates a cobra.Command
// which may interact with the host.
func PachctlCmd(address string) (*cobra.Command, error) {
	var verbose bool
	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Long: `Access the Pachyderm API.

Environment variables (and defaults):
  ADDRESS=0.0.0.0:30650, the server to connect to.
`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !verbose {
				// Silence any grpc logs
				grpclog.SetLogger(log.New(ioutil.Discard, "", 0))
				// Silence our FUSE logs
				lion.SetLevel(lion.LevelNone)
			}
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Output verbose logs")

	pfsCmds := pfscmds.Cmds(address)
	for _, cmd := range pfsCmds {
		rootCmd.AddCommand(cmd)
	}
	ppsCmds, err := ppscmds.Cmds(address)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	for _, cmd := range ppsCmds {
		rootCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(deploycmds.DeployCmd())

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
				return sanitizeErr(err)
			}
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			version, err := versionClient.GetVersion(ctx, &google_protobuf.Empty{})

			if err != nil {
				fmt.Fprintf(writer, "pachd\t(version unknown) : error connecting to pachd server at address (%v): %v\n\nplease make sure pachd is up (`kubectl get all`) and portforwarding is enabled\n", address, sanitizeErr(err))
				return writer.Flush()
			}

			printVersion(writer, "pachd", version)
			return writer.Flush()
		}),
	}
	deleteAll := &cobra.Command{
		Use:   "delete-all",
		Short: "Delete everything.",
		Long: `Delete all repos, commits, files, pipelines and jobs.
This resets the cluster to its initial state.`,
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewFromAddress(address)
			if err != nil {
				return sanitizeErr(err)
			}
			fmt.Printf("Are you sure you want to delete all repos, commits, files, pipelines and jobs? yN\n")
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
	var port int
	portForward := &cobra.Command{
		Use:   "port-forward",
		Short: "Forward a port on the local machine to pachd. This command blocks.",
		Long:  "Forward a port on the local machine to pachd. This command blocks.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl get pod -l app=pachd |  awk '{if (NR!=1) { print $1; exit 0 }}')
kubectl port-forward "$pod" %d:650
`, port))
			fmt.Println("Port forwarded, CTRL-C to exit.")
			return pkgexec.RunIO(pkgexec.IO{
				Stdin:  stdin,
				Stderr: os.Stderr,
			}, "sh")
		}),
	}
	portForward.Flags().IntVarP(&port, "port", "p", 30650, "The local port to bind to.")
	rootCmd.AddCommand(version)
	rootCmd.AddCommand(deleteAll)
	rootCmd.AddCommand(portForward)
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

func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(grpc.ErrorDesc(err))
}
