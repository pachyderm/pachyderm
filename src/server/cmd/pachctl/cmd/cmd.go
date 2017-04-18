package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/client/version/versionpb"
	pfscmds "github.com/pachyderm/pachyderm/src/server/pfs/cmds"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	deploycmds "github.com/pachyderm/pachyderm/src/server/pkg/deploy/cmds"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	ppscmds "github.com/pachyderm/pachyderm/src/server/pps/cmds"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

// PachctlCmd takes a pachd host-address and creates a cobra.Command
// which may interact with the host.
func PachctlCmd(address string) (*cobra.Command, error) {
	var verbose bool
	var noMetrics bool
	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Long: `Access the Pachyderm API.

Environment variables:
  ADDRESS=<host>:<port>, the pachd server to connect to (e.g. 127.0.0.1:30650).
`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !verbose {
				// Silence any grpc logs
				l := log.New()
				l.Level = log.FatalLevel
				grpclog.SetLogger(l)
			}
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Output verbose logs")
	rootCmd.PersistentFlags().BoolVarP(&noMetrics, "no-metrics", "", false, "Don't report user metrics for this command")

	pfsCmds := pfscmds.Cmds(address, &noMetrics)
	for _, cmd := range pfsCmds {
		rootCmd.AddCommand(cmd)
	}
	ppsCmds, err := ppscmds.Cmds(address, &noMetrics)
	if err != nil {
		return nil, sanitizeErr(err)
	}
	for _, cmd := range ppsCmds {
		rootCmd.AddCommand(cmd)
	}
	deployCmds := deploycmds.Cmds(&noMetrics)
	for _, cmd := range deployCmds {
		rootCmd.AddCommand(cmd)
	}

	version := &cobra.Command{
		Use:   "version",
		Short: "Return version information.",
		Long:  "Return version information.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			if !noMetrics {
				metricsFn := metrics.ReportAndFlushUserAction("Version")
				defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			printVersionHeader(writer)
			printVersion(writer, "pachctl", version.Version)
			writer.Flush()

			versionClient, err := getVersionAPIClient(address)
			if err != nil {
				return sanitizeErr(err)
			}
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			version, err := versionClient.GetVersion(ctx, &types.Empty{})

			if err != nil {
				buf := bytes.NewBufferString("")
				errWriter := tabwriter.NewWriter(buf, 20, 1, 3, ' ', 0)
				fmt.Fprintf(errWriter, "pachd\t(version unknown) : error connecting to pachd server at address (%v): %v\n\nplease make sure pachd is up (`kubectl get all`) and portforwarding is enabled\n", address, sanitizeErr(err))
				errWriter.Flush()
				return errors.New(buf.String())
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
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewMetricsClientFromAddress(address, !noMetrics, "user")
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
	var kubeCtlFlags string
	portForward := &cobra.Command{
		Use:   "port-forward",
		Short: "Forward a port on the local machine to pachd. This command blocks.",
		Long:  "Forward a port on the local machine to pachd. This command blocks.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl %v get pod -l app=pachd | awk '{if (NR!=1) { print $1; exit 0 }}')
kubectl %v port-forward "$pod" %d:650
`, kubeCtlFlags, kubeCtlFlags, port))
			fmt.Println("Port forwarded, CTRL-C to exit.")
			return cmdutil.RunIO(cmdutil.IO{
				Stdin:  stdin,
				Stderr: os.Stderr,
			}, "sh")
		}),
	}
	portForward.Flags().IntVarP(&port, "port", "p", 30650, "The local port to bind to.")
	portForward.Flags().StringVarP(&kubeCtlFlags, "kubectlflags", "k", "", "Any kubectl flags to proxy, e.g. --kubectlflags='--kubeconfig /some/path/kubeconfig'")

	var uiPort int
	var uiWebsocketPort int
	var uiKubeCtlFlags string
	portForwardUI := &cobra.Command{
		Use:   "port-forward-ui",
		Short: "Forward a port on the local machine to pachyderm dash. This command blocks.",
		Long:  "Forward a port on the local machine to pachyderm dash. This command blocks.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			var eg errgroup.Group

			eg.Go(func() error {
				stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl %v get pod -l app=dash | awk '{if (NR!=1) { print $1; exit 0 }}')
kubectl %v port-forward "$pod" %d:8080
`, uiKubeCtlFlags, uiKubeCtlFlags, uiPort))
				fmt.Printf("Dash UI Port forwarded, navigate to localhost:%v, CTRL-C to exit.", uiPort)
				return cmdutil.RunIO(cmdutil.IO{
					Stdin:  stdin,
					Stderr: os.Stderr,
				}, "sh")
			})

			eg.Go(func() error {
				stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl %v get pod -l app=dash | awk '{if (NR!=1) { print $1; exit 0 }}')
kubectl %v port-forward "$pod" %d:8081
`, uiKubeCtlFlags, uiKubeCtlFlags, uiWebsocketPort))
				fmt.Println("Websocket Port forwarded")
				return cmdutil.RunIO(cmdutil.IO{
					Stdin:  stdin,
					Stderr: os.Stderr,
				}, "sh")
			})

			return eg.Wait()
		}),
	}
	portForwardUI.Flags().IntVarP(&uiPort, "port", "p", 38080, "The local port to bind to.")
	portForwardUI.Flags().IntVarP(&uiWebsocketPort, "proxy-port", "x", 32082, "The local port to bind to.")
	portForwardUI.Flags().StringVarP(&uiKubeCtlFlags, "kubectlflags", "k", "", "Any kubectl flags to proxy, e.g. --kubectlflags='--kubeconfig /some/path/kubeconfig'")

	rootCmd.AddCommand(version)
	rootCmd.AddCommand(deleteAll)
	rootCmd.AddCommand(portForward)
	rootCmd.AddCommand(portForwardUI)
	return rootCmd, nil
}

func getVersionAPIClient(address string) (versionpb.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return versionpb.NewAPIClient(clientConn), nil
}

func printVersionHeader(w io.Writer) {
	fmt.Fprintf(w, "COMPONENT\tVERSION\t\n")
}

func printVersion(w io.Writer, component string, v *versionpb.Version) {
	fmt.Fprintf(w, "%s\t%s\t\n", component, version.PrettyPrintVersion(v))
}

func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(grpc.ErrorDesc(err))
}
