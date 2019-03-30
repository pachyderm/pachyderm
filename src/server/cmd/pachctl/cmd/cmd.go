package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"text/tabwriter"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
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
	logutil "github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	ppscmds "github.com/pachyderm/pachyderm/src/server/pps/cmds"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	bashCompletionFunc = `
__pachctl_get_object() {
	if pachctl_output=$(eval pachctl $1 2>/dev/null); then
		local out=($(echo "${pachctl_output}" | awk -v c=$2 'NR > 1 {print $c}'))
		COMPREPLY+=($(compgen -P "${__pachctl_prefix}" -S "${__pachctl_suffix}" -W "${out[*]}" "$cur"))
	fi
}

__pachctl_get_repo() {
	__pachctl_get_object "list-repo" 1
}

# $1: repo name
__pachctl_get_commit() {
	if [[ -z $1 ]]; then
		return
	fi
	__pachctl_get_object "list-commit $1" 2
	__pachctl_get_object "list-branch $1" 1
}

# Performs completion of the standard format <repo>@<branch-or-commit>
__pachctl_get_repo_commit() {
	if [[ ${cur} != *@* ]]; then
		compopt -o nospace
		local __pachctl_suffix="@"
		__pachctl_get_repo
	else
		local repo=$(__parse_repo ${cur})
		local cur=$(__parse_commit ${cur})
		local __pachctl_prefix="${repo}@"
		__pachctl_get_commit ${repo}
	fi
}

# Performs completion of the standard format <repo>@<branch> (will not match commits)
__pachctl_get_repo_branch() {
	if [[ ${cur} != *@* ]]; then
		compopt -o nospace
		local __pachctl_suffix="@"
		__pachctl_get_repo
	else
		local repo=$(__parse_repo ${cur})
		local cur=$(__parse_commit ${cur})
		local __pachctl_prefix="${repo}@"
		__pachctl_get_branch ${repo}
	fi
}

# Performs completion of the standard format <repo>@<branch-or-commit>:<path>
__pachctl_get_repo_commit_path() {
	# completion treats ':' as its own argument and an argument-break
	if [[ ${#nouns[@]} -ge 1 ]] && [[ ${cur} == : ]]; then
		local repo=$(__parse_repo ${nouns[-1]})
		local commit=$(__parse_commit ${nouns[-1]})
		local cur=
		__pachctl_get_path ${repo} ${commit}
	elif [[ ${#nouns[@]} -ge 2 ]] && [[ ${nouns[-1]} == : ]]; then
		local repo=$(__parse_repo ${nouns[-2]})
		local commit=$(__parse_commit ${nouns[-2]})
		__pachctl_get_path ${repo} ${commit}
	elif [[ ${cur} != *@* ]]; then
		__pachctl_get_repo_commit
	else
		compopt -o nospace
		local __pachctl_suffix=":"
		__pachctl_get_repo_commit
	fi
}

# $1: repo name
# $2: branch name or commit id
__pachctl_get_path() {
	if [[ -z $1 ]] || [[ -z $2 ]]; then
		return
	fi

	__pachctl_get_object "glob-file \"$1@$2:${cur}**\"" 2
}

# $1: repo name
__pachctl_get_branch() {
	if [[ -z $1 ]]; then
		return
	fi
	__pachctl_get_object "list-branch $1" 1
}

__pachctl_get_job() {
	__pachctl_get_object "list-job" 1
}

__pachctl_get_pipeline() {
	__pachctl_get_object "list-pipeline" 1
}

__pachctl_get_datum() {
	__pachctl_get_object "list-datum $2" 1
}

# Parses repo from the format <repo>@<branch-or-commit>
__parse_repo() {
	echo $1 | cut -f 1 -d "@"
}

# Parses commit from the format <repo>@<branch-or-commit>
__parse_commit() {
	echo $1 | cut -f 2 -d "@" -s
}

__pachctl_auth_scope() {
	local out=(none reader writer owner)
	COMPREPLY+=($(compgen -W "${out[*]}" "$cur"))
}

# $*: an integer corresponding to a positional argument index (zero-based)
# Checks if the argument being completed is a positional argument at one of the
# specified indices.
__is_active_arg() {
	for predicate in $*; do
		if [[ $predicate == ${#nouns[@]} ]]; then
			return 0
		fi
	done
	return 1
}

__custom_func() {
	case ${last_command} in
		pachctl_auth_check)
		  if __is_active_arg 0; then
				__pachctl_auth_scope
			elif __is_active_arg 1; then
				__pachctl_get_repo
			fi
			;;
		pachctl_auth_get)
			if __is_active_arg 0 1; then
				__pachctl_get_repo
			fi
			;;
		pachctl_auth_set)
		  if __is_active_arg 1; then
			  __pachctl_auth_scope
			elif __is_active_arg 2; then
				__pachctl_get_repo
			fi
			;;
		pachctl_update-repo | pachctl_inspect-repo | pachctl_delete-repo | pachctl_list-branch | pachctl_list-commit)
			if __is_active_arg 0; then
				__pachctl_get_repo
			fi
			;;
		pachctl_delete-branch | pachctl_subscribe-commit)
			if __is_active_arg 0; then
				__pachctl_get_repo_branch
			fi
			;;
		pachctl_finish-commit | pachctl_inspect-commit | pachctl_delete-commit | pachctl_create-branch | pachctl_start-commit)
			if __is_active_arg 0; then
				__pachctl_get_repo_commit
			fi
			;;
		pachctl_set-branch)
			if __is_active_arg 0; then
				__pachctl_get_repo_branch 0
			elif __is_active_arg 1; then
				__pachctl_get_branch $(__parse_repo ${nouns[0]})
			fi
			;;
		pachctl_get-file | pachctl_inspect-file | pachctl_list-file | pachctl_delete-file | pachctl_glob-file | pachctl_put-file)
			# completion splits the ':' character into its own argument
			if __is_active_arg 0 1 2; then
				__pachctl_get_repo_commit_path
			fi
			;;
		pachctl_copy-file | pachctl_diff-file)
			__pachctl_get_repo_commit_path
			;;
		pachctl_inspect-job | pachctl_delete-job | pachctl_stop-job | pachctl_list-datum | pachctl_restart-datum)
			if __is_active_arg 0; then
				__pachctl_get_job
			fi
			;;
		pachctl_inspect-datum)
			if __is_active_arg 0; then
				__pachctl_get_job
			elif __is_active_arg 1; then
				__pachctl_get_datum ${nouns[0]}
			fi
			;;
		pachctl_inspect-pipeline | pachctl_delete-pipeline | pachctl_start-pipeline | pachctl_stop-pipeline | pachctl_extract-pipeline | pachctl_edit-pipeline)
			if __is_active_arg 0; then
				__pachctl_get_pipeline
			fi
			;;
		pachctl_flush-job | pachctl_flush-commit)
			__pachctl_get_repo_commit
			;;
		*)
			;;
	esac
}`
)

// PachctlCmd creates a cobra.Command which can deploy pachyderm clusters and
// interact with them (it implements the pachctl binary).
func PachctlCmd() *cobra.Command {
	var verbose bool
	var noMetrics bool
	var noPortForwarding bool

	raw := false
	rawFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	rawFlags.BoolVar(&raw, "raw", false, "disable pretty printing, print raw json")

	marshaller := &jsonpb.Marshaler{Indent: "  "}

	rootCmd := &cobra.Command{
		Use:  os.Args[0],
		Long: `Access the Pachyderm API.

Environment variables:
  PACHD_ADDRESS=<host>:<port>, the pachd server to connect to (e.g. 127.0.0.1:30650).
  PACH_CONFIG=<path>, the path where pachctl will attempt to load your pach config.
  JAEGER_ENDPOINT=<host>:<port>, the Jaeger server to connect to, if PACH_ENABLE_TRACING is set
  PACH_ENABLE_TRACING={true,false}, If true, and JAEGER_ENDPOINT is set, attach a
    Jaeger trace to all outgoing RPCs
`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetFormatter(new(prefixed.TextFormatter))

			if !verbose {
				log.SetLevel(log.ErrorLevel)
				// Silence grpc logs
				grpclog.SetLoggerV2(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard))
			} else {
				log.SetLevel(log.DebugLevel)
				// etcd overrides grpc's logs--there's no way to enable one without
				// enabling both.
				// Error and warning logs are discarded because they will be
				// redundantly sent to the info logger. See:
				// https://godoc.org/google.golang.org/grpc/grpclog#NewLoggerV2
				logger := log.StandardLogger()
				etcd.SetLogger(grpclog.NewLoggerV2(
					logutil.NewGRPCLogWriter(logger, "etcd/grpc"),
					ioutil.Discard,
					ioutil.Discard,
				))
			}
		},
		BashCompletionFunction: bashCompletionFunc,
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Output verbose logs")
	rootCmd.PersistentFlags().BoolVarP(&noMetrics, "no-metrics", "", false, "Don't report user metrics for this command")
	rootCmd.PersistentFlags().BoolVarP(&noPortForwarding, "no-port-forwarding", "", false, "Disable implicit port forwarding")

	var subcommands []*cobra.Command

	var clientOnly bool
	var timeoutFlag string
	versionCmd := &cobra.Command{
		Short: "Return version information.",
		Long:  "Return version information.",
		Run:   cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
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
				pachClient, err = client.NewOnUserMachine(false, !noPortForwarding, "user", client.WithDialTimeout(timeout))
			} else {
				pachClient, err = client.NewOnUserMachine(false, !noPortForwarding, "user")
			}
			if err != nil {
				return err
			}
			defer pachClient.Close()
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
	versionCmd.Flags().StringVar(&timeoutFlag, "timeout", "default", "If set, "+
		"'pachctl version' will timeout after the given duration (formatted as a "+
		"golang time duration--a number followed by ns, us, ms, s, m, or h). If "+
		"--client-only is set, this flag is ignored. If unset, pachctl will use a "+
		"default timeout; if set to 0s, the call will never time out.")
	versionCmd.Flags().AddFlagSet(rawFlags)
	subcommands = append(subcommands, cmdutil.CreateAliases(versionCmd, []string{"version"})...)

	deleteAll := &cobra.Command{
		Short: "Delete everything.",
		Long:  `Delete all repos, commits, files, pipelines and jobs.
This resets the cluster to its initial state.`,
		Run:   cmdutil.RunFixedArgs(0, func(args []string) error {
			client, err := client.NewOnUserMachine(!noMetrics, !noPortForwarding, "user")
			if err != nil {
				return err
			}
			defer client.Close()
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
			fmt.Println("All ACLs, repos, commits, files, pipelines and jobs will be deleted.")
			if len(repos) > 0 {
				fmt.Printf("Repos to delete: %s\n", strings.Join(repos, ", "))
			}
			if len(pipelines) > 0 {
				fmt.Printf("Pipelines to delete: %s\n", strings.Join(pipelines, ", "))
			}
			fmt.Println("Are you sure you want to do this? (y/n):")
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
	subcommands = append(subcommands, cmdutil.CreateAliases(deleteAll, []string{"delete all"})...)

	var port uint16
	var remotePort uint16
	var samlPort uint16
	var uiPort uint16
	var uiWebsocketPort uint16
	var pfsPort uint16
	var namespace string
	portForward := &cobra.Command{
		Short: "Forward a port on the local machine to pachd. This command blocks.",
		Long:  "Forward a port on the local machine to pachd. This command blocks.",
		Run:   cmdutil.RunFixedArgs(0, func(args []string) error {
			fw, err := client.NewPortForwarder(namespace)
			if err != nil {
				return err
			}

			if err = fw.Lock(); err != nil {
				return err
			}

			defer fw.Close()

			failCount := 0

			fmt.Println("Forwarding the pachd (Pachyderm daemon) port...")
			if err = fw.RunForDaemon(port, remotePort); err != nil {
				fmt.Printf("%v\n", err)
				failCount++
			}

			fmt.Println("Forwarding the SAML ACS port...")
			if err = fw.RunForSAMLACS(samlPort); err != nil {
				fmt.Printf("%v\n", err)
				failCount++
			}

			fmt.Printf("Forwarding the dash (Pachyderm dashboard) UI port to http://localhost:%v...\n", uiPort)
			if err = fw.RunForDashUI(uiPort); err != nil {
				fmt.Printf("%v\n", err)
				failCount++
			}

			fmt.Println("Forwarding the dash (Pachyderm dashboard) websocket port...")
			if err = fw.RunForDashWebSocket(uiWebsocketPort); err != nil {
				fmt.Printf("%v\n", err)
				failCount++
			}

			fmt.Println("Forwarding the PFS port...")
			if err = fw.RunForPFS(pfsPort); err != nil {
				fmt.Printf("%v\n", err)
				failCount++
			}

			if failCount < 5 {
				fmt.Println("CTRL-C to exit")
				fmt.Println("NOTE: kubernetes port-forward often outputs benign error messages, these should be ignored unless they seem to be impacting your ability to connect over the forwarded port.")

				ch := make(chan os.Signal, 1)
				signal.Notify(ch, os.Interrupt)
				<-ch
			}

			return nil
		}),
	}
	portForward.Flags().Uint16VarP(&port, "port", "p", 30650, "The local port to bind pachd to.")
	portForward.Flags().Uint16Var(&remotePort, "remote-port", 650, "The remote port that pachd is bound to in the cluster.")
	portForward.Flags().Uint16Var(&samlPort, "saml-port", 30654, "The local port to bind pachd's SAML ACS to.")
	portForward.Flags().Uint16VarP(&uiPort, "ui-port", "u", 30080, "The local port to bind Pachyderm's dash service to.")
	portForward.Flags().Uint16VarP(&uiWebsocketPort, "proxy-port", "x", 30081, "The local port to bind Pachyderm's dash proxy service to.")
	portForward.Flags().Uint16VarP(&pfsPort, "pfs-port", "f", 30652, "The local port to bind PFS over HTTP to.")
	portForward.Flags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace Pachyderm is deployed in.")
	subcommands = append(subcommands, cmdutil.CreateAliases(portForward, []string{"port-forward"})...)

	var install bool
	var path string
	completion := &cobra.Command{
		Short: "Print or install the bash completion code.",
		Long:  "Print or install the bash completion code. This should be placed as the file `pachctl` in the bash completion directory (by default this is `/etc/bash_completion.d`. If bash-completion was installed via homebrew, this would be `$(brew --prefix)/etc/bash_completion.d`.)",
		Run:   cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			var dest io.Writer

			if install {
				f, err := os.Create(path)

				if err != nil {
					if os.IsPermission(err) {
						fmt.Fprintf(os.Stderr, "Permission error installing completions, rerun this command with sudo.\n")
					}
					return err
				}

				defer func() {
					if err := f.Close(); err != nil && retErr == nil {
						retErr = err
					}

					fmt.Printf("Bash completions installed in %s, you must restart bash to enable completions.\n", path)
				}()

				dest = f
			} else {
				dest = os.Stdout
			}

			return rootCmd.GenBashCompletion(dest)
		}),
	}
	completion.Flags().BoolVar(&install, "install", false, "Install the completion.")
	completion.Flags().StringVar(&path, "path", "/etc/bash_completion.d/pachctl", "Path to install the completion to. This will default to `/etc/bash_completion.d/` if unspecified.")
	subcommands = append(subcommands, cmdutil.CreateAliases(completion, []string{"completion"})...)

	// Logical commands for grouping commands by verb (no run functions)
	deleteDocs := &cobra.Command{
		Short: "Delete an existing Pachyderm resource.",
		Long:  "Delete an existing Pachyderm resource.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(deleteDocs, []string{"delete"})...)

	createDocs := &cobra.Command{
		Short: "Create a new instance of a Pachyderm resource.",
		Long:  "Create a new instance of a Pachyderm resource.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(createDocs, []string{"create"})...)

	updateDocs := &cobra.Command{
		Short: "Change the properties of an existing Pachyderm resource.",
		Long:  "Change the properties of an existing Pachyderm resource.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(updateDocs, []string{"update"})...)

	inspectDocs := &cobra.Command{
		Short: "Show detailed information about a Pachyderm resource.",
		Long:  "Show detailed information about a Pachyderm resource.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(inspectDocs, []string{"inspect"})...)

	listDocs := &cobra.Command{
		Short: "Print a list of Pachyderm resources of a specific type.",
		Long:  "Print a list of Pachyderm resources of a specific type.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(listDocs, []string{"list"})...)

	startDocs := &cobra.Command{
		Short: "Start a Pachyderm resource.",
		Long:  "Start a Pachyderm resource.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(startDocs, []string{"start"})...)

	finishDocs := &cobra.Command{
		Short: "Finish a Pachyderm resource.",
		Long:  "Finish a Pachyderm resource.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(finishDocs, []string{"finish"})...)

	flushDocs := &cobra.Command{
		Short: "Wait for the side-effects of a Pachyderm resource to propagate.",
		Long:  "Wait for the side-effects of a Pachyderm resource to propagate.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(flushDocs, []string{"flush"})...)

	subscribeDocs := &cobra.Command{
		Short: "Wait for notifications of changes to a Pachyderm resource.",
		Long:  "Wait for notifications of changes to a Pachyderm resource.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(subscribeDocs, []string{"subscribe"})...)

	putDocs := &cobra.Command{
		Short: "Insert data into the Pachyderm filesystem.",
		Long:  "Insert data into the Pachyderm filesystem.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(putDocs, []string{"put"})...)

	copyDocs := &cobra.Command{
		Short: "Copy data between locations in the Pachyderm filesystem",
		Long:  "Copy data between locations in the Pachyderm filesystem",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(copyDocs, []string{"copy"})...)

	getDocs := &cobra.Command{
		Short: "Get the raw data represented by a Pachyderm resource.",
		Long:  "Get the raw data represented by a Pachyderm resource.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(getDocs, []string{"get"})...)

	globDocs := &cobra.Command{
		Short: "Print a list of Pachyderm resources matching a glob pattern.",
		Long:  "Print a list of Pachyderm resources matching a glob pattern.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(globDocs, []string{"glob"})...)

	diffDocs := &cobra.Command{
		Short: "Show the differences between two Pachyderm resources.",
		Long:  "Show the differences between two Pachyderm resources.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(diffDocs, []string{"diff"})...)

	stopDocs := &cobra.Command{
		Short: "Cancel an ongoing task.",
		Long:  "Cancel an ongoing task.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(stopDocs, []string{"stop"})...)

	restartDocs := &cobra.Command{
		Short: "Resume a stopped task.",
		Long:  "Resume a stopped task.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(restartDocs, []string{"restart"})...)

	editDocs := &cobra.Command{
		Short: "Edit the value of an existing Pachyderm resource.",
		Long:  "Edit the value of an existing Pachyderm resource.",
	}
	subcommands = append(subcommands, cmdutil.CreateAliases(editDocs, []string{"edit"})...)

	subcommands = append(subcommands, pfscmds.Cmds(&noMetrics, &noPortForwarding)...)
	subcommands = append(subcommands, ppscmds.Cmds(&noMetrics, &noPortForwarding)...)
	subcommands = append(subcommands, deploycmds.Cmds(&noMetrics, &noPortForwarding)...)
	subcommands = append(subcommands, authcmds.Cmds(&noMetrics, &noPortForwarding)...)
	subcommands = append(subcommands, enterprisecmds.Cmds(&noMetrics, &noPortForwarding)...)
	subcommands = append(subcommands, admincmds.Cmds(&noMetrics, &noPortForwarding)...)
	subcommands = append(subcommands, debugcmds.Cmds(&noMetrics, &noPortForwarding)...)

	cmdutil.MergeCommands(rootCmd, subcommands)

	applyRootUsageFunc(rootCmd)
	applyCommandCompat1_8(rootCmd, &noMetrics, &noPortForwarding)

	return rootCmd
}

func printVersionHeader(w io.Writer) {
	fmt.Fprintf(w, "COMPONENT\tVERSION\t\n")
}

func printVersion(w io.Writer, component string, v *versionpb.Version) {
	fmt.Fprintf(w, "%s\t%s\t\n", component, version.PrettyPrintVersion(v))
}

func applyRootUsageFunc(rootCmd *cobra.Command) {
	/*
	`Usage:{{if .Runnable}}
  {{if .HasAvailableFlags}}{{appendIfNotPresent .UseLine "[flags]"}}{{else}}{{.UseLine}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{ if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{ if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableSubCommands }}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
	*/

	originalUsageFunc := rootCmd.UsageFunc()
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		if cmd != rootCmd {
			return originalUsageFunc(cmd)
		}

		return originalUsageFunc(cmd)
	})
}
