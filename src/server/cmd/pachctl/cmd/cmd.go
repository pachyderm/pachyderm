package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/facebookgo/pidfile"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/client/version/versionpb"
	authcmds "github.com/pachyderm/pachyderm/src/server/auth/cmds"
	enterprisecmds "github.com/pachyderm/pachyderm/src/server/enterprise/cmds"
	pfscmds "github.com/pachyderm/pachyderm/src/server/pfs/cmds"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	deploycmds "github.com/pachyderm/pachyderm/src/server/pkg/deploy/cmds"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	ppscmds "github.com/pachyderm/pachyderm/src/server/pps/cmds"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	bashCompletionPath = "/etc/bash_completion.d/pachctl"
)

// PachctlCmd creates a cobra.Command which can deploy pachyderm clusters and
// interact with them (it implements the pachctl binary).
func PachctlCmd() (*cobra.Command, error) {
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

	pfsCmds := pfscmds.Cmds(&noMetrics)
	for _, cmd := range pfsCmds {
		rootCmd.AddCommand(cmd)
	}
	ppsCmds, err := ppscmds.Cmds(&noMetrics)
	if err != nil {
		return nil, err
	}
	for _, cmd := range ppsCmds {
		rootCmd.AddCommand(cmd)
	}
	deployCmds := deploycmds.Cmds(&noMetrics)
	for _, cmd := range deployCmds {
		rootCmd.AddCommand(cmd)
	}
	authCmds := authcmds.Cmds()
	for _, cmd := range authCmds {
		rootCmd.AddCommand(cmd)
	}
	enterpriseCmds := enterprisecmds.Cmds()
	for _, cmd := range enterpriseCmds {
		rootCmd.AddCommand(cmd)
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Return version information.",
		Long:  "Return version information.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			if !noMetrics {
				start := time.Now()
				startMetricsWait := metrics.StartReportAndFlushUserAction("Version", start)
				defer startMetricsWait()
				defer func() {
					finishMetricsWait := metrics.FinishReportAndFlushUserAction("Version", retErr, start)
					finishMetricsWait()
				}()
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			printVersionHeader(writer)
			printVersion(writer, "pachctl", version.Version)
			writer.Flush()

			cfg, err := config.Read()
			if err != nil {
				log.Warningf("error loading user config from ~/.pachderm/config: %v", err)
			}
			pachdAddress := client.GetAddressFromUserMachine(cfg)
			versionClient, err := getVersionAPIClient(pachdAddress)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			version, err := versionClient.GetVersion(ctx, &types.Empty{})

			if err != nil {
				buf := bytes.NewBufferString("")
				errWriter := tabwriter.NewWriter(buf, 20, 1, 3, ' ', 0)
				fmt.Fprintf(errWriter, "pachd\t(version unknown) : error connecting to pachd server at address (%v): %v\n\nplease make sure pachd is up (`kubectl get all`) and portforwarding is enabled\n", pachdAddress, grpcutil.ScrubGRPC(err))
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
			client, err := client.NewOnUserMachine(!noMetrics, "user")
			if err != nil {
				return err
			}
			fmt.Printf("Are you sure you want to delete all ACLs, repos, commits, files, pipelines and jobs? yN\n")
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
	var uiPort int
	var uiWebsocketPort int
	var kubeCtlFlags string
	portForward := &cobra.Command{
		Use:   "port-forward",
		Short: "Forward a port on the local machine to pachd. This command blocks.",
		Long:  "Forward a port on the local machine to pachd. This command blocks.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			pidfile.SetPidfilePath(path.Join(os.Getenv("HOME"), ".pachyderm/port-forward.pid"))
			pid, err := pidfile.Read()
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			if pid != 0 {
				if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
					if !strings.Contains(err.Error(), "no such process") {
						return err
					}
				}
			}
			if err := pidfile.Write(); err != nil {
				return err
			}

			var eg errgroup.Group

			eg.Go(func() error {
				stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl %v get pod -l app=pachd | awk '{if (NR!=1) { print $1; exit 0 }}')
kubectl %v port-forward "$pod" %d:650
`, kubeCtlFlags, kubeCtlFlags, port))
				return cmdutil.RunIO(cmdutil.IO{
					Stdin:  stdin,
					Stderr: os.Stderr,
				}, "sh")
			})

			eg.Go(func() error {
				stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl %v get pod -l app=dash | awk '{if (NR!=1) { print $1; exit 0 }}')
kubectl %v port-forward "$pod" %d:8080
`, kubeCtlFlags, kubeCtlFlags, uiPort))
				if err := cmdutil.RunIO(cmdutil.IO{
					Stdin: stdin,
				}, "sh"); err != nil {
					return fmt.Errorf("UI not enabled, deploy with --dashboard")
				}
				return nil
			})

			eg.Go(func() error {
				stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl %v get pod -l app=dash | awk '{if (NR!=1) { print $1; exit 0 }}')
kubectl %v port-forward "$pod" %d:8081
`, kubeCtlFlags, kubeCtlFlags, uiWebsocketPort))
				cmdutil.RunIO(cmdutil.IO{
					Stdin: stdin,
				}, "sh")
				return nil
			})

			fmt.Printf("Pachd port forwarded\nDash websocket port forwarded\nDash UI port forwarded, navigate to localhost:%v\nCTRL-C to exit", uiPort)
			return eg.Wait()
		}),
	}
	portForward.Flags().IntVarP(&port, "port", "p", 30650, "The local port to bind to.")
	portForward.Flags().IntVarP(&uiPort, "ui-port", "u", 30080, "The local port to bind to.")
	portForward.Flags().IntVarP(&uiWebsocketPort, "proxy-port", "x", 30081, "The local port to bind to.")
	portForward.Flags().StringVarP(&kubeCtlFlags, "kubectlflags", "k", "", "Any kubectl flags to proxy, e.g. --kubectlflags='--kubeconfig /some/path/kubeconfig'")

	garbageCollect := &cobra.Command{
		Use:   "garbage-collect",
		Short: "Garbage collect unused data.",
		Long: `Garbage collect unused data.

When a file/commit/repo is deleted, the data is not immediately removed from the underlying storage system (e.g. S3) for performance and architectural reasons.  This is similar to how when you delete a file on your computer, the file is not necessarily wiped from disk immediately.

To actually remove the data, you will need to manually invoke garbage collection.  The easiest way to do it is through "pachctl garbage-collecth".

Currently "pachctl garbage-collect" can only be started when there are no active jobs running.  You also need to ensure that there's no ongoing "put-file".  Garbage collection puts the cluster into a readonly mode where no new jobs can be created and no data can be added.
`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			client, err := client.NewOnUserMachine(!noMetrics, "user")
			if err != nil {
				return err
			}

			return client.GarbageCollect()
		}),
	}

	var from, to, namespace string
	migrate := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the internal state of Pachyderm from one version to another.",
		Long: `Migrate the internal state of Pachyderm from one version to
another.  Note that most of the time updating Pachyderm doesn't
require a migration.  Refer to the docs for your specific version
to find out if it requires a migration.

It's highly recommended that you only run migrations when there are no
activities in your cluster, e.g. no jobs should be running.

The migration command takes the general form:

$ pachctl migrate --from <FROM_VERSION> --to <TO_VERSION>

If "--from" is not provided, pachctl will attempt to discover the current
version of the cluster.  If "--to" is not provided, pachctl will use the
version of pachctl itself.

Example:

# Migrate Pachyderm from 1.4.8 to 1.5.0
$ pachctl migrate --from 1.4.8 --to 1.5.0
`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			// If `from` is not provided, we use the cluster version.
			if from == "" {
				cfg, err := config.Read()
				if err != nil {
					log.Warningf("error loading user config from ~/.pachderm/config: %v", err)
				}
				pachdAddress := client.GetAddressFromUserMachine(cfg)
				versionClient, err := getVersionAPIClient(pachdAddress)
				if err != nil {
					return err
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				clusterVersion, err := versionClient.GetVersion(ctx, &types.Empty{})
				if err != nil {
					return fmt.Errorf("unable to discover cluster version; please provide the --from flag.  Error: %v", err)
				}
				from = version.PrettyPrintVersionNoAdditional(clusterVersion)
			}

			// if `to` is not provided, we use the version of pachctl itself.
			if to == "" {
				to = version.PrettyPrintVersionNoAdditional(version.Version)
			}

			jobSpec := batch.Job{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Job",
					APIVersion: "batch/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "pach-migration",
					Labels: map[string]string{
						"suite": "pachyderm",
					},
				},
				Spec: batch.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:    "migration",
									Image:   fmt.Sprintf("pachyderm/pachd:%v", version.PrettyPrintVersion(version.Version)),
									Command: []string{"/pachd", fmt.Sprintf("--migrate=%v-%v", from, to)},
								},
							},
							RestartPolicy: "OnFailure",
						},
					},
				},
			}

			tmpFile, err := ioutil.TempFile("", "")
			if err != nil {
				return err
			}
			defer os.Remove(tmpFile.Name())

			encoder := json.NewEncoder(tmpFile)
			encoder.Encode(jobSpec)
			tmpFile.Close()

			cmd := exec.Command("kubectl", "create", "--validate=false", "-f", tmpFile.Name())
			out, err := cmd.CombinedOutput()
			fmt.Println(string(out))
			if err != nil {
				return err
			}
			fmt.Println("Successfully launched migration.  To see the progress, use `kubectl logs job/pach-migration`")
			return nil
		}),
	}
	migrate.Flags().StringVar(&from, "from", "", "The current version of the cluster.  If not specified, pachctl will attempt to discover the version of the cluster.")
	migrate.Flags().StringVar(&to, "to", "", "The version of Pachyderm to migrate to.  If not specified, pachctl will use its own version.")
	migrate.Flags().StringVar(&namespace, "namespace", "default", "The kubernetes namespace under which Pachyderm is deployed.")

	completion := &cobra.Command{
		Use:   "completion",
		Short: "Install bash completion code.",
		Long:  "Install bash completion code.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			bashCompletionFile, err := os.Create(bashCompletionPath)
			if err != nil {
				if os.IsPermission(err) {
					fmt.Fprintf(os.Stderr, "Permission error installing completions, rerun this command with sudo:\n$ sudo env \"PATH=$PATH\" pachctl completion\n")
				}
				return err
			}
			defer func() {
				if err := bashCompletionFile.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			if err := rootCmd.GenBashCompletion(bashCompletionFile); err != nil {
				return err
			}
			fmt.Printf("Bash completions installed in %s, you must restart bash to enable completions.\n", bashCompletionPath)
			return nil
		}),
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(deleteAll)
	rootCmd.AddCommand(portForward)
	rootCmd.AddCommand(garbageCollect)
	rootCmd.AddCommand(migrate)
	rootCmd.AddCommand(completion)
	return rootCmd, nil
}

func getVersionAPIClient(address string) (versionpb.APIClient, error) {
	clientConn, err := grpc.Dial(address, client.PachDialOptions()...)
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
