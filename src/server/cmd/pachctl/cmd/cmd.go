package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	golog "log"
	"os"
	"path"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/facebookgo/pidfile"
	"github.com/fatih/color"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/client/version/versionpb"
	admincmds "github.com/pachyderm/pachyderm/src/server/admin/cmds"
	authcmds "github.com/pachyderm/pachyderm/src/server/auth/cmds"
	debugcmds "github.com/pachyderm/pachyderm/src/server/debug/cmds"
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
	bashCompletionFunc = `
__pachctl_get_object() {
	if [[ ${#nouns[@]} -ne $1 ]]; then
		return
	fi

	local pachctl_output out
	if pachctl_output=$(eval pachctl $2 2>/dev/null); then
		out=($(echo "${pachctl_output}" | awk -v c=$3 'NR > 1 {print $c}'))
		COMPREPLY+=($(compgen -W "${out[*]}" -- "$cur"))
	fi
}

__pachctl_get_repo() {
	__pachctl_get_object $1 "list-repo" 1
}

__pachctl_get_commit() {
	__pachctl_get_object $1 "list-commit $2" 2
}

__pachctl_get_path() {
	__pachctl_get_object $1 "glob-file $2 $3 \"${words[${#words[@]}-1]}**\"" 1
}

__pachctl_get_branch() {
	__pachctl_get_object $1 "list-branch $2" 1
}

__pachctl_get_job() {
	__pachctl_get_object $1 "list-job" 1
}

__pachctl_get_pipeline() {
	__pachctl_get_object $1 "list-pipeline" 1
}

__pachctl_get_datum() {
	__pachctl_get_object $1 "list-datum $2" 1
}

__custom_func() {
	case ${last_command} in
		pachctl_update-repo | pachctl_inspect-repo | pachctl_delete-repo | pachctl_list-commit | pachctl_list-branch)
			__pachctl_get_repo 0
			;;
		pachctl_start-commit | pachctl_subscribe-commit | pachctl_delete-branch)
			__pachctl_get_repo 0
			__pachctl_get_branch 1 ${nouns[0]}
			;;
		pachctl_finish-commit | pachctl_inspect-commit | pachctl_delete-commit | pachctl_glob-file)
			__pachctl_get_repo 0
			__pachctl_get_commit 1 ${nouns[0]}
			;;
		pachctl_set-branch)
			__pachctl_get_repo 0
			__pachctl_get_commit 1 ${nouns[0]}
			__pachctl_get_branch 1 ${nouns[0]}
			;;
		pachctl_put-file)
			__pachctl_get_repo 0
			__pachctl_get_branch 1 ${nouns[0]}
			__pachctl_get_path 2 ${nouns[0]} ${nouns[1]}
			;;
		pachctl_copy-file | pachctl_diff-file)
			__pachctl_get_repo 0
			__pachctl_get_commit 1 ${nouns[0]}
			__pachctl_get_path 2 ${nouns[0]} ${nouns[1]}
			__pachctl_get_repo 3
			__pachctl_get_commit 4 ${nouns[3]}
			__pachctl_get_path 5 ${nouns[3]} ${nouns[4]}
			;;
		pachctl_get-file | pachctl_inspect-file | pachctl_list-file | pachctl_delete-file)
			__pachctl_get_repo 0
			__pachctl_get_commit 1 ${nouns[0]}
			__pachctl_get_path 2 ${nouns[0]} ${nouns[1]}
			;;
		pachctl_inspect-job | pachctl_delete-job | pachctl_stop-job | pachctl_list-datum)
			__pachctl_get_job 0
			;;
		pachctl_inspect-datum)
			__pachctl_get_job 0
			__pachctl_get_datum 1 ${nouns[0]}
			;;
		pachctl_inspect-pipeline | pachctl_delete-pipeline | pachctl_start-pipeline | pachctl_stop-pipeline | pachctl_run-pipeline)
			__pachctl_get_pipeline 0
			;;
		*)
			;;
	esac
}`
)

type logWriter golog.Logger

func (l *logWriter) Write(p []byte) (int, error) {
	err := (*golog.Logger)(l).Output(2, string(p))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// PachctlCmd creates a cobra.Command which can deploy pachyderm clusters and
// interact with them (it implements the pachctl binary).
func PachctlCmd() (*cobra.Command, error) {
	var verbose bool
	var noMetrics bool
	raw := false
	rawFlag := func(cmd *cobra.Command) {
		cmd.Flags().BoolVar(&raw, "raw", false, "disable pretty printing, print raw json")
	}
	marshaller := &jsonpb.Marshaler{Indent: "  "}

	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Long: `Access the Pachyderm API.

Environment variables:
  ADDRESS=<host>:<port>, the pachd server to connect to (e.g. 127.0.0.1:30650).
`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !verbose {
				// Silence grpc logs
				l := log.New()
				l.Level = log.FatalLevel
				grpclog.SetLoggerV2(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard))
			} else {
				// etcd overrides grpc's logs--there's no way to enable one without
				// enabling both
				etcd.SetLogger(grpclog.NewLoggerV2(
					(*logWriter)(golog.New(os.Stderr, "[etcd/grpc] INFO  ", golog.LstdFlags|golog.Lshortfile)),
					(*logWriter)(golog.New(os.Stderr, "[etcd/grpc] WARN  ", golog.LstdFlags|golog.Lshortfile)),
					(*logWriter)(golog.New(os.Stderr, "[etcd/grpc] ERROR ", golog.LstdFlags|golog.Lshortfile)),
				))
			}

		},
		BashCompletionFunction: bashCompletionFunc,
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
	adminCmds := admincmds.Cmds(&noMetrics)
	for _, cmd := range adminCmds {
		rootCmd.AddCommand(cmd)
	}
	debugCmds := debugcmds.Cmds(&noMetrics)
	for _, cmd := range debugCmds {
		rootCmd.AddCommand(cmd)
	}

	var clientOnly bool
	var timeoutFlag string
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Return version information.",
		Long:  "Return version information.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			if clientOnly {
				if raw {
					if err := marshaller.Marshal(os.Stdout, version.Version); err != nil {
						return err
					}
				} else {
					fmt.Println(version.PrettyPrintVersion(version.Version))
				}
				return nil
			}
			if !noMetrics {
				start := time.Now()
				startMetricsWait := metrics.StartReportAndFlushUserAction("Version", start)
				defer startMetricsWait()
				defer func() {
					finishMetricsWait := metrics.FinishReportAndFlushUserAction("Version", retErr, start)
					finishMetricsWait()
				}()
			}

			// Print header + client version
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			if raw {
				if err := marshaller.Marshal(os.Stdout, version.Version); err != nil {
					return err
				}
			} else {
				printVersionHeader(writer)
				printVersion(writer, "pachctl", version.Version)
				if err := writer.Flush(); err != nil {
					return err
				}
			}

			// Dial pachd & get server version
			var pachClient *client.APIClient
			var err error
			if timeoutFlag != "default" {
				var timeout time.Duration
				timeout, err = time.ParseDuration(timeoutFlag)
				if err != nil {
					return fmt.Errorf("could not parse timeout duration %q: %v", timeout, err)
				}
				pachClient, err = client.NewOnUserMachine(false, "user", client.WithDialTimeout(timeout))
			} else {
				pachClient, err = client.NewOnUserMachine(false, "user")
			}
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			version, err := pachClient.GetVersion(ctx, &types.Empty{})

			if err != nil {
				buf := bytes.NewBufferString("")
				errWriter := tabwriter.NewWriter(buf, 20, 1, 3, ' ', 0)
				fmt.Fprintf(errWriter, "pachd\t(version unknown) : error connecting to pachd server at address (%v): %v\n\nplease make sure pachd is up (`kubectl get all`) and portforwarding is enabled\n", pachClient.GetAddress(), grpc.ErrorDesc(err))
				errWriter.Flush()
				return errors.New(buf.String())
			}

			// print server version
			if raw {
				if err := marshaller.Marshal(os.Stdout, version); err != nil {
					return err
				}
			} else {
				printVersion(writer, "pachd", version)
				if err := writer.Flush(); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	versionCmd.Flags().BoolVar(&clientOnly, "client-only", false, "If set, "+
		"only print pachctl's version, but don't make any RPCs to pachd. Useful "+
		"if pachd is unavailable")
	rawFlag(versionCmd)
	versionCmd.Flags().StringVar(&timeoutFlag, "timeout", "default", "If set, "+
		"pachctl version will timeout after the given duration (formatted as a "+
		"golang time duration--a number followed by ns, us, ms, s, m, or h). If "+
		"--client-only is set, this flag is ignored. If unset, pachctl will use a "+
		"default timeout; if set to 0s, the call will never time out.")

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
			red := color.New(color.FgRed).SprintFunc()
			var repos, pipelines []string
			repoInfos, err := client.ListRepo()
			if err != nil {
				return err
			}
			for _, ri := range repoInfos {
				repos = append(repos, red(ri.Repo.Name))
			}
			pipelineInfos, err := client.ListPipeline()
			if err != nil {
				return err
			}
			for _, pi := range pipelineInfos {
				pipelines = append(pipelines, red(pi.Pipeline.Name))
			}
			fmt.Printf("Are you sure you want to delete all ACLs, repos, commits, files, pipelines and jobs?\nyN\n")
			if len(repos) > 0 {
				fmt.Printf("Repos to delete: %s\n", strings.Join(repos, ", "))
			}
			if len(pipelines) > 0 {
				fmt.Printf("Pipelines to delete: %s\n", strings.Join(pipelines, ", "))
			}
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
	var samlPort int
	var uiPort int
	var uiWebsocketPort int
	var kubeCtlFlags string
	var namespace string
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

			kubeCtlFlags += fmt.Sprintf(" --namespace=%s", namespace)

			var eg errgroup.Group

			eg.Go(func() error {
				fmt.Println("Forwarding the pachd (Pachyderm daemon) port...")
				stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl %s get pod -l app=pachd  --output='jsonpath={.items[0].metadata.name}')
kubectl %s port-forward "$pod" %d:650
`, kubeCtlFlags, kubeCtlFlags, port))
				if err := cmdutil.RunIO(cmdutil.IO{
					Stdin:  stdin,
					Stderr: os.Stderr,
				}, "sh"); err != nil {
					fmt.Fprintln(os.Stderr, "Could not forward Pachd port")
					return fmt.Errorf("Could not forward Pachd port")
				}
				return nil
			})

			eg.Go(func() error {
				fmt.Println("Forwarding the SAML ACS port...")
				stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl %s get pod -l suite=pachyderm,app=pachd  --output='jsonpath={.items[0].metadata.name}')
kubectl %s port-forward "$pod" %d:654
`, kubeCtlFlags, kubeCtlFlags, samlPort))
				if err := cmdutil.RunIO(cmdutil.IO{
					Stdin:  stdin,
					Stderr: os.Stderr,
				}, "sh"); err != nil {
					fmt.Fprintln(os.Stderr, "Could not forward Pachyderm SAML ACS port")
					return fmt.Errorf("Could not forward Pachyderm SAML ACS port")
				}
				return nil
			})

			eg.Go(func() error {
				stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl %s get pod -l app=dash --output='jsonpath={.items[0].metadata.name}')
kubectl %s port-forward "$pod" %d:8080
`, kubeCtlFlags, kubeCtlFlags, uiPort))
				if err := cmdutil.RunIO(cmdutil.IO{
					Stdin: stdin,
				}, "sh"); err != nil {
					fmt.Fprintln(os.Stderr, "Is the dashboard deployed? If not, deploy with \"pachctl deploy local --dashboard-only\"")
					return fmt.Errorf("Could not forward dash UI port")
				}
				return nil
			})

			eg.Go(func() error {
				fmt.Printf("Forwarding the dash (Pachyderm dashboard) UI port to http://localhost:%v ...\n", uiPort)
				stdin := strings.NewReader(fmt.Sprintf(`
pod=$(kubectl %v get pod -l app=dash --output='jsonpath={.items[0].metadata.name}')
kubectl %v port-forward "$pod" %d:8081
`, kubeCtlFlags, kubeCtlFlags, uiWebsocketPort))
				if err := cmdutil.RunIO(cmdutil.IO{
					Stdin: stdin,
				}, "sh"); err != nil {
					fmt.Fprintln(os.Stderr, "Could not forward dash websocket port")
					return fmt.Errorf("Could not forward dash websocket port")
				}
				return nil
			})

			fmt.Println("CTRL-C to exit")
			fmt.Println("NOTE: kubernetes port-forward often outputs benign error messages, these should be ignored unless they seem to be impacting your ability to connect over the forwarded port.")

			return eg.Wait()
		}),
	}
	portForward.Flags().IntVarP(&port, "port", "p", 30650, "The local port to bind pachd to.")
	portForward.Flags().IntVar(&samlPort, "saml-port", 30654, "The local port to bind pachd's SAML ACS to.")
	portForward.Flags().IntVarP(&uiPort, "ui-port", "u", 30080, "The local port to bind Pachyderm's dash service to.")
	portForward.Flags().IntVarP(&uiWebsocketPort, "proxy-port", "x", 30081, "The local port to bind Pachyderm's dash proxy service to.")
	portForward.Flags().StringVarP(&kubeCtlFlags, "kubectlflags", "k", "", "Any kubectl flags to proxy, e.g. --kubectlflags='--kubeconfig /some/path/kubeconfig'")
	portForward.Flags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace Pachyderm is deployed in.")

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
	rootCmd.AddCommand(completion)
	return rootCmd, nil
}

func printVersionHeader(w io.Writer) {
	fmt.Fprintf(w, "COMPONENT\tVERSION\t\n")
}

func printVersion(w io.Writer, component string, v *versionpb.Version) {
	fmt.Fprintf(w, "%s\t%s\t\n", component, version.PrettyPrintVersion(v))
}
