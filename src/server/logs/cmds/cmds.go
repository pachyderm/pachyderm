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

func Cmds(pachCtx *config.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
	var (
		commands        []*cobra.Command
		logQL, pipeline string
		project         = pachCtx.Project
		datum           string
		app             string
		from            = cmdutil.TimeFlag(time.Now().Add(-700 * time.Hour))
		to              = cmdutil.TimeFlag(time.Now())
		offset          uint
	)
	logsCmd := &cobra.Command{
		// TODO(CORE-2200): remove references to “new.”
		Short: "New logs functionality",
		Long:  "Query Pachyderm using new log service.",
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
			req.Filter = new(logs.LogFilter)
			req.Filter.TimeRange = &logs.TimeRangeLogFilter{
				From:   timestamppb.New(time.Time(from)),
				Until:  timestamppb.New(time.Time(to)),
				Offset: uint64(offset),
			}
			switch {
			case cmd.Flag("logql").Changed:
				if cmd.Flag("project").Changed || cmd.Flag("pipeline").Changed || cmd.Flag("datum").Changed || cmd.Flag("app").Changed {
					fmt.Fprintln(os.Stderr, "only one of [--logQL | --project PROJECT --pipeline PIPELINE | --datum DATUM] may be set")
					os.Exit(1)
				}
				addLogQLRequest(req, logQL)
			case cmd.Flag("pipeline").Changed:
				if cmd.Flag("datum").Changed || cmd.Flag("app").Changed {
					fmt.Fprintln(os.Stderr, "only one of [--logQL | --project PROJECT --pipeline PIPELINE | --datum DATUM] may be set")
					os.Exit(1)
				}
				addPipelineRequest(req, project, pipeline)
			case cmd.Flag("project").Changed:
				if cmd.Flag("datum").Changed || cmd.Flag("app").Changed {
					fmt.Fprintln(os.Stderr, "only one of [--logQL | --project PROJECT --pipeline PIPELINE | --datum DATUM] may be set")
					os.Exit(1)
				}
				addProjectRequest(req, project)
			case cmd.Flag("datum").Changed:
				if cmd.Flag("app").Changed {
					fmt.Fprintln(os.Stderr, "only one of [--logQL | --project PROJECT --pipeline PIPELINE | --datum DATUM] may be set")
					os.Exit(1)
				}
				addDatumRequest(req, datum)
			case cmd.Flag("app").Changed:
				addLogQLRequest(req, fmt.Sprintf(`{app=%q}`, app))
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
	logsCmd.Flags().StringVar(&datum, "datum", datum, "Datum for datum query.")
	logsCmd.Flags().StringVar(&app, "app", app, "App for service query.")
	logsCmd.Flags().Var(&from, "from", "Return logs at or after this time.")
	logsCmd.Flags().Var(&to, "to", "Return logs before  this time.")
	logsCmd.Flags().UintVar(&offset, "offset", offset, "Number of logs to skip at beginning of time range.")
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
