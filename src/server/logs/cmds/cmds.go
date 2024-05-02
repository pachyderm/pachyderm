package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alessio/shellescape"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/logs"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
)

func isAdmin(ctx context.Context, client *client.APIClient) (bool, error) {
	authResp, err := client.AuthAPIClient.GetPermissions(ctx, &auth.GetPermissionsRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}})
	if err != nil {
		if status.Code(err) != codes.Unimplemented {
			return false, err
		}
		return true, nil
	}
	for _, role := range authResp.Roles {
		if role == auth.ClusterAdminRole {
			return true, nil
		}
	}
	return false, nil
}

func addLogQLRequest(req *logs.GetLogsRequest, logQL string) {
	req.Query = &logs.LogQuery{
		QueryType: &logs.LogQuery_Admin{
			Admin: &logs.AdminLogQuery{
				AdminType: &logs.AdminLogQuery_Logql{
					Logql: logQL,
				},
			},
		},
	}
}

func addPipelineRequest(req *logs.GetLogsRequest, project, pipeline string) {
	req.Query = &logs.LogQuery{
		QueryType: &logs.LogQuery_User{
			User: &logs.UserLogQuery{
				UserType: &logs.UserLogQuery_Pipeline{
					Pipeline: &logs.PipelineLogQuery{
						Project:  project,
						Pipeline: pipeline,
					},
				},
			},
		},
	}
}

func addProjectRequest(req *logs.GetLogsRequest, project string) {
	req.Query = &logs.LogQuery{
		QueryType: &logs.LogQuery_User{
			User: &logs.UserLogQuery{
				UserType: &logs.UserLogQuery_Project{
					Project: project,
				},
			},
		},
	}
}

func addJobDatumRequest(req *logs.GetLogsRequest, job, datum string) {
	req.Query = &logs.LogQuery{
		QueryType: &logs.LogQuery_User{
			User: &logs.UserLogQuery{
				UserType: &logs.UserLogQuery_JobDatum{
					JobDatum: &logs.JobDatumLogQuery{
						Job:   job,
						Datum: datum,
					},
				},
			},
		},
	}
}

func addDatumRequest(req *logs.GetLogsRequest, datum string) {
	req.Query = &logs.LogQuery{
		QueryType: &logs.LogQuery_User{
			User: &logs.UserLogQuery{
				UserType: &logs.UserLogQuery_Datum{
					Datum: datum,
				},
			},
		},
	}
}

func addPodRequest(req *logs.GetLogsRequest, pod string) {
	req.Query = &logs.LogQuery{
		QueryType: &logs.LogQuery_Admin{
			Admin: &logs.AdminLogQuery{
				AdminType: &logs.AdminLogQuery_Pod{
					Pod: pod,
				},
			},
		},
	}
}

func addPodContainerRequest(req *logs.GetLogsRequest, pod, container string) {
	req.Query = &logs.LogQuery{
		QueryType: &logs.LogQuery_Admin{
			Admin: &logs.AdminLogQuery{
				AdminType: &logs.AdminLogQuery_PodContainer{
					PodContainer: &logs.PodContainer{
						Container: container,
						Pod:       pod,
					},
				},
			},
		},
	}
}

func addAppRequest(req *logs.GetLogsRequest, app string) {
	req.Query = &logs.LogQuery{
		QueryType: &logs.LogQuery_Admin{
			Admin: &logs.AdminLogQuery{
				AdminType: &logs.AdminLogQuery_App{
					App: app,
				},
			},
		},
	}
}

func hasDisallowedFlags(cmd *cobra.Command, flags ...string) bool {
	for _, f := range flags {
		if cmd.Flag(f).Changed {
			return true
		}
	}
	return false
}

func Cmds(pachCtx *config.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
	var (
		commands  []*cobra.Command
		logQL     string
		project   = pachCtx.Project
		pipeline  string
		job       string
		datum     string
		from      = cmdutil.TimeFlag(time.Now().Add(-700 * time.Hour))
		to        = cmdutil.TimeFlag(time.Now())
		offset    uint
		pod       string
		container string
		app       string
		limit     uint
	)
	logsCmd := &cobra.Command{
		// TODO(CORE-2200): Remove references to “new” and unhide.
		Hidden: true,
		Short:  "New logs functionality",
		Long:   "Query Pachyderm using new log service.",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer client.Close()

			isAdmin, err := isAdmin(client.Ctx(), client)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			var req = new(logs.GetLogsRequest)
			req.Filter = &logs.LogFilter{
				Limit: uint64(limit),
			}
			req.Filter.TimeRange = &logs.TimeRangeLogFilter{
				From:   timestamppb.New(time.Time(from)),
				Until:  timestamppb.New(time.Time(to)),
				Offset: uint64(offset),
			}
			switch {
			case cmd.Flag("logql").Changed:
				if hasDisallowedFlags(cmd, "project", "pipeline", "datum", "app", "job") {
					fmt.Fprintln(os.Stderr, "only one of [--logQL | --project PROJECT --pipeline PIPELINE | --datum DATUM] may be set")
					os.Exit(1)
				}
				addLogQLRequest(req, logQL)
			case cmd.Flag("pipeline").Changed:
				if hasDisallowedFlags(cmd, "datum", "app") {
					fmt.Fprintln(os.Stderr, "only one of [--logQL | --project PROJECT --pipeline PIPELINE | --datum DATUM] may be set")
					os.Exit(1)
				}
				addPipelineRequest(req, project, pipeline)
			case cmd.Flag("project").Changed:
				if hasDisallowedFlags(cmd, "datum", "app") {
					fmt.Fprintln(os.Stderr, "only one of [--logQL | --project PROJECT --pipeline PIPELINE | --datum DATUM] may be set")
					os.Exit(1)
				}
				addProjectRequest(req, project)
			case cmd.Flag("datum").Changed && cmd.Flag("job").Changed:
				if hasDisallowedFlags(cmd, "app") {
					fmt.Fprintln(os.Stderr, "only one of [--logQL | --project PROJECT --pipeline PIPELINE | --datum DATUM] may be set")
					os.Exit(1)
				}
				addJobDatumRequest(req, job, datum)
			case cmd.Flag("datum").Changed:
				if hasDisallowedFlags(cmd, "app") {
					fmt.Fprintln(os.Stderr, "only one of [--logQL | --project PROJECT --pipeline PIPELINE | --datum DATUM] may be set")
					os.Exit(1)
				}
				addDatumRequest(req, datum)
			case !cmd.Flag("pod").Changed && cmd.Flag("container").Changed:
				if !isAdmin {
					fmt.Fprintln(os.Stderr, "only users with the ClusterAdmin role can view logs by pod.")
					os.Exit(1)
				}
				fmt.Fprintln(os.Stderr, "to specify --container, specifying --pod is required.")
				os.Exit(1)
			case cmd.Flag("pod").Changed && cmd.Flag("container").Changed:
				if !isAdmin {
					fmt.Fprintln(os.Stderr, "only users with the ClusterAdmin role can view logs by pod.")
					os.Exit(1)
				}
				addPodContainerRequest(req, pod, container)
			case cmd.Flag("pod").Changed:
				if !isAdmin {
					fmt.Fprintln(os.Stderr, "only users with the ClusterAdmin role can view logs by pod.")
					os.Exit(1)
				}
				addPodRequest(req, pod)
			case cmd.Flag("app").Changed:
				if !isAdmin {
					fmt.Fprintln(os.Stderr, "only users with the ClusterAdmin role can view logs by pod.")
					os.Exit(1)
				}
				addAppRequest(req, app)
			case isAdmin:
				addLogQLRequest(req, `{suite="pachyderm"}`)
			default:
				addLogQLRequest(req, `{pod=~".+"}`)
			}

			// always ask for paging hints
			req.WantPagingHint = true
			resp, err := client.LogsClient.GetLogs(client.Ctx(), req)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			for {
				resp, err := resp.Recv()
				if err != nil {
					if err != io.EOF {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
					break
				}

				switch resp.ResponseType.(type) {
				case *logs.GetLogsResponse_Log:
					fmt.Println(string(resp.GetLog().GetVerbatim().GetLine()))
				case *logs.GetLogsResponse_PagingHint:
					hint := resp.GetPagingHint()
					if hint == nil {
						fmt.Fprintf(os.Stderr, "ERROR: do not know how to handle %v\n`", resp)
						continue
					}
					// printing to stderr in order to keep stdout clean
					fmt.Fprintln(os.Stderr, toPachctl(cmd, args, hint))
				default:
					fmt.Fprintf(os.Stderr, "ERROR: do not know how to handle %T\n`", resp)
				}
			}
		},
		Use: "logs2",
	}
	logsCmd.Flags().StringVar(&logQL, "logql", "", "LogQL query")
	logsCmd.Flags().StringVar(&project, "project", project, "Project for pipeline query.")
	logsCmd.Flags().StringVar(&pipeline, "pipeline", pipeline, "Pipeline for pipeline query.")
	logsCmd.Flags().StringVar(&job, "job", job, "Job for job query.")
	logsCmd.Flags().StringVar(&datum, "datum", datum, "Datum for datum query.")
	logsCmd.Flags().Var(&from, "from", "Return logs at or after this time.")
	logsCmd.Flags().Var(&to, "to", "Return logs before  this time.")
	logsCmd.Flags().UintVar(&offset, "offset", offset, "Number of logs to skip at beginning of time range.")
	logsCmd.Flags().StringVar(&pod, "pod", pod, "Pod in the cluster.")
	logsCmd.Flags().UintVar(&limit, "limit", limit, "Maximum number of logs to return (0 for unlimited).")
	logsCmd.Flags().StringVar(&container, "container", container, "Container name belonging to the pod specified in the --pod argument.")
	logsCmd.Flags().StringVar(&app, "app", app, "Return logs for all pods with a certain value for the label 'app'.")
	commands = append(commands, logsCmd)
	return commands
}

func toFlags(flags map[string]string, hint *logs.GetLogsRequest) string {
	var result string

	if from := hint.GetFilter().GetTimeRange().GetFrom(); !from.AsTime().IsZero() {
		result += " --from " + shellescape.Quote(from.AsTime().Format(time.RFC3339Nano))
	}
	if until := hint.GetFilter().GetTimeRange().GetUntil(); !until.AsTime().IsZero() {
		result += " --to " + shellescape.Quote(until.AsTime().Format(time.RFC3339Nano))
	}
	if offset := hint.GetFilter().GetTimeRange().GetOffset(); offset != 0 {
		result += fmt.Sprintf(" --offset %d", offset)
	}
	for flag, arg := range flags {
		if flag == "from" || flag == "to" || flag == "offset" {
			continue
		}
		result += " --" + flag + " " + shellescape.Quote(arg)
	}
	return result
}

func toPachctl(cmd *cobra.Command, args []string, hint *logs.PagingHint) string {
	var flags = make(map[string]string)
	older := cmd.CommandPath()
	newer := cmd.CommandPath()
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if !cmd.Flag(flag.Name).Changed {
			return
		}
		flags[flag.Name] = flag.Value.String()
	})

	older += toFlags(flags, hint.GetOlder())
	newer += toFlags(flags, hint.GetNewer())
	for _, arg := range args {
		older += " " + arg
		newer += " " + arg
	}
	return fmt.Sprintf("Older logs are available with %s\nNewer logs are available with %s", older, newer)
}
