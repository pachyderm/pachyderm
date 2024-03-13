package cmds

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/logs"
)

func Cmds(ctx context.Context, pachCtx *config.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	var logQL string
	logsCmd := &cobra.Command{
		// TODO(QQQ): remove references to “new.”
		Short: "New logs functionality",
		Long:  "Query Pachyderm using new log service.",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer client.Close()

			authResp, err := client.AuthAPIClient.GetPermissions(client.Ctx(), &auth.GetPermissionsRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}})
			if err != nil {
				os.Exit(1)
			}
			var isAdmin bool
			for _, role := range authResp.Roles {
				if role == auth.ClusterAdminRole {
					isAdmin = true
				}
			}

			var req = new(logs.GetLogsRequest)
			switch {
			case logQL != "":
				if isAdmin {
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
			default:
				if isAdmin {
					req.Query = &logs.LogQuery{
						QueryType: &logs.LogQuery_Admin{
							Admin: &logs.AdminLogQuery{
								AdminType: &logs.AdminLogQuery_Logql{
									Logql: `{suite="pachyderm"}`,
								},
							},
						},
					}
				} else {
					req.Query = &logs.LogQuery{
						QueryType: &logs.LogQuery_Admin{
							Admin: &logs.AdminLogQuery{
								AdminType: &logs.AdminLogQuery_Logql{
									Logql: `{}`,
								},
							},
						},
					}
				}
			}

			resp, err := client.LogsClient.GetLogs(client.Ctx(), req)
			for {
				resp, err := resp.Recv()
				if err != nil {
					if err != io.EOF {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
					break
				}
				switch log := resp.GetLog().GetLogType().(type) {
				case *logs.LogMessage_PpsLogMessage:
					fmt.Println(resp.GetLog().GetPpsLogMessage())
				case *logs.LogMessage_Json:
					fmt.Println(resp.GetLog().GetJson().GetVerbatim().GetLine())
				case *logs.LogMessage_Verbatim:
					fmt.Println(string(resp.GetLog().GetVerbatim().GetLine()))
				default:
					fmt.Fprintf(os.Stderr, "ERROR: do not know how to handle %T\n`", log)
				}

			}
		},
		Use: "logs2",
	}
	logsCmd.Flags().StringVar(&logQL, "logql", "", "LogQL query")
	commands = append(commands, logsCmd)
	return commands
}
