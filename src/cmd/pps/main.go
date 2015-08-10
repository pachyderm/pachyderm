package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm"
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

	inspectCmd := cobramainutil.Command{
		Use:        "inspect github.com/user/repository [path/to/specDir]",
		Long:       "Inspect a pipeline specification.",
		MinNumArgs: 1,
		MaxNumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			pipelineArgs, err := getPipelineArgs(args)
			if err != nil {
				return err
			}
			getPipelineResponse, err := ppsutil.GetPipelineGithub(
				apiClient,
				pipelineArgs.contextDir,
				pipelineArgs.user,
				pipelineArgs.repository,
				pipelineArgs.branch,
				pipelineArgs.accessToken,
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
			pipelineArgs, err := getPipelineArgs(args)
			if err != nil {
				return err
			}
			startPipelineRunResponse, err := ppsutil.StartPipelineRunGithub(
				apiClient,
				pipelineArgs.contextDir,
				pipelineArgs.user,
				pipelineArgs.repository,
				pipelineArgs.branch,
				pipelineArgs.accessToken,
			)
			if err != nil {
				return err
			}
			fmt.Println(startPipelineRunResponse.PipelineRunId)
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
			fmt.Printf("%s %s\n", args[0], strings.Replace(name, "PIPELINE_RUN_STATUS_TYPE_", "", -1))
			return nil
		},
	}.ToCobraCommand()

	rootCmd := &cobra.Command{
		Use: "pps",
		Long: `Access the PPS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PPS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:651.`,
	}
	rootCmd.AddCommand(cobramainutil.NewVersionCommand(clientConn, pachyderm.Version))
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	return rootCmd.Execute()
}

type pipelineArgs struct {
	contextDir  string
	user        string
	repository  string
	branch      string
	accessToken string
}

func getPipelineArgs(args []string) (*pipelineArgs, error) {
	path := args[0]
	if !strings.HasPrefix(path, "github.com/") {
		return nil, fmt.Errorf("%s is not supported", path)
	}
	split := strings.Split(path, "/")
	if len(split) != 3 {
		return nil, fmt.Errorf("%s is not supported", path)
	}
	contextDir := ""
	if len(args) > 1 {
		contextDir = args[1]
	}
	return &pipelineArgs{
		contextDir:  contextDir,
		user:        split[1],
		repository:  split[2],
		branch:      "",
		accessToken: "",
	}, nil
}
