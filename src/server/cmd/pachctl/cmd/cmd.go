package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
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

func kubectlGetAddress() (string, error) {
	stdout := bytes.NewBuffer(nil)
	err := pkgexec.RunIO(pkgexec.IO{Stdout: stdout, Stderr: os.Stderr},
		"kubectl", "get", "nodes", "-o=jsonpath={$.items[0].status.addresses[?(@.type==\"LegacyHostIP\")].address}")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not get IP of kubernetes node from kubectl: %s", err))
	}
	return fmt.Sprintf("%s:30650", stdout.String()), nil
}

// PachctlCmd takes a pachd host-address and creates a cobra.Command
// which may interact with the host.
func PachctlCmd(address string) (*cobra.Command, error) {
	var verbose bool
	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Long: `Access the Pachyderm API.

Environment variables (and defaults):
	ADDRESS=<host>:<port>, the server to connect to (by default, <host> is node 0 in 'kubectl get nodes', and <port> is 30650)
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

	if address == "" {
		var err error
		address, err = kubectlGetAddress()
		if err != nil {
			return nil, err
		}
	}

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

func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(grpc.ErrorDesc(err))
}
