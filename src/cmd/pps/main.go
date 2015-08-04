package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pkg/cobramainutil"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/ppsutil"
	"github.com/spf13/cobra"
)

var (
	defaultEnv = map[string]string{
		"PPS_ADDRESS": "0.0.0.0:651",
	}
)

type appEnv struct {
	Address string `env:"PPS_ADDRESS"`
}

func main() {
	mainutil.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)

	clientConn, err := grpc.Dial(appEnv.Address)
	if err != nil {
		return err
	}
	apiClient := pps.NewApiClient(clientConn)

	var protoFlag bool

	versionCmd := cobramainutil.Command{
		Use:  "version",
		Long: "Print the version.",
		Run: func(cmd *cobra.Command, args []string) error {
			getVersionResponse, err := ppsutil.GetVersion(apiClient)
			if err != nil {
				return err
			}
			fmt.Printf("Client: %s\nServer: %s\n", common.VersionString(), pps.VersionString(getVersionResponse.Version))
			return nil
		},
	}.ToCobraCommand()

	inspectCmd := cobramainutil.Command{
		Use:        "inspect github.com/user/repository [path/to/specDir]",
		Long:       "Inspect a pipeline specification.",
		MinNumArgs: 1,
		MaxNumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			if !strings.HasPrefix(path, "github.com/") {
				return fmt.Errorf("%s is not supported", path)
			}
			split := strings.Split(path, "/")
			if len(split) != 3 {
				return fmt.Errorf("%s is not supported", path)
			}
			branch := ""
			accessToken := ""
			contextDir := ""
			if len(args) > 1 {
				contextDir = args[1]
			}
			getPipelineResponse, err := ppsutil.GetPipelineGithub(
				apiClient,
				contextDir,
				split[1],
				split[2],
				branch,
				accessToken,
			)
			if err != nil {
				return err
			}
			if protoFlag {
				fmt.Printf("%v\n", getPipelineResponse.Pipeline)
			} else {
				data, err := json.MarshalIndent(getPipelineResponse.Pipeline, "", "\t ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			}
			return nil
		},
	}.ToCobraCommand()
	inspectCmd.Flags().BoolVar(&protoFlag, "proto", false, "Print in proto format instead of JSON.")

	startCmd := cobramainutil.Command{
		Use:        "start github.com/user/repository [path/to/specDir]",
		Long:       "Start a pipeline specification run.",
		MinNumArgs: 1,
		MaxNumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			if !strings.HasPrefix(path, "github.com/") {
				return fmt.Errorf("%s is not supported", path)
			}
			split := strings.Split(path, "/")
			if len(split) != 3 {
				return fmt.Errorf("%s is not supported", path)
			}
			branch := ""
			accessToken := ""
			contextDir := ""
			if len(args) > 1 {
				contextDir = args[1]
			}
			startPipelineRunResponse, err := ppsutil.StartPipelineRunGithub(
				apiClient,
				contextDir,
				split[1],
				split[2],
				branch,
				accessToken,
			)
			if err != nil {
				return err
			}
			fmt.Printf("started pipeline run with id %s\n", startPipelineRunResponse.PipelineRunId)
			return nil
		},
	}.ToCobraCommand()

	statusCmd := cobramainutil.Command{
		Use:     "status pipelineRunID",
		Long:    "Get the status of a pipeline run.",
		NumArgs: 1,
		Run: func(cmd *cobra.Command, args []string) error {
			getPipelineRunStatusResponse, err := ppsutil.GetPipelineRunStatus(
				apiClient,
				args[0],
			)
			if err != nil {
				return err
			}
			name, ok := pps.PipelineRunStatusType_name[int32(getPipelineRunStatusResponse.PipelineRunStatus.PipelineRunStatusType)]
			if !ok {
				return fmt.Errorf("unknown run status")
			}
			fmt.Printf("%s: %s\n", args[0], strings.Replace(name, "PIPELINE_RUN_STATUS_TYPE_", "", -1))
			return nil
		},
	}.ToCobraCommand()

	rootCmd := &cobra.Command{
		Use: "pps",
		Long: `Access the PPS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PPS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:651.`,
	}
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	return rootCmd.Execute()
}
