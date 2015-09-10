package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.pedge.io/proto/client"
	"go.pedge.io/proto/time"

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
	PachydermPpsd1Port string `env:"PACHYDERM_PPSD_1_PORT"`
	Address            string `env:"PPS_ADDRESS"`
}

func main() {
	mainutil.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)

	address := appEnv.PachydermPpsd1Port
	if address == "" {
		address = appEnv.Address
	} else {
		address = strings.Replace(address, "tcp://", "", -1)
	}
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
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
				"",
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
				"",
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

	logsCmd := cobramainutil.Command{
		Use:        "logs pipelineRunID [node]",
		Long:       "Get the logs from a pipeline run.",
		MinNumArgs: 1,
		MaxNumArgs: 2,
		Run: func(cmd *cobra.Command, args []string) error {
			node := ""
			if len(args) == 2 {
				node = args[1]
			}
			getPipelineRunLogsResponse, err := ppsutil.GetPipelineRunLogs(
				apiClient,
				args[0],
				node,
			)
			if err != nil {
				return err
			}
			for _, pipelineRunLog := range getPipelineRunLogsResponse.PipelineRunLog {
				name, ok := pps.OutputStream_name[int32(pipelineRunLog.OutputStream)]
				if !ok {
					return fmt.Errorf("unknown pps.OutputStream")
				}
				name = strings.Replace(name, "OUTPUT_STREAM_", "", -1)
				containerID := pipelineRunLog.ContainerId
				if len(containerID) > 8 {
					containerID = containerID[:8]
				}
				logInfo := &logInfo{
					Node:         pipelineRunLog.Node,
					ContainerID:  containerID,
					OutputStream: name,
					Time:         prototime.TimestampToTime(pipelineRunLog.Timestamp),
				}
				logInfoData, err := json.Marshal(logInfo)
				if err != nil {
					return err
				}
				s := fmt.Sprintf("%s %s", string(logInfoData), string(pipelineRunLog.Data))
				fmt.Println(strings.TrimSpace(s))
			}
			return nil
		},
	}.ToCobraCommand()

	rootCmd := &cobra.Command{
		Use: "pps",
		Long: `Access the PPS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PPS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:651.`,
	}
	rootCmd.AddCommand(protoclient.NewVersionCommand(clientConn, pachyderm.Version, nil))
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
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

type logInfo struct {
	Node         string
	ContainerID  string
	Time         time.Time
	OutputStream string
}
