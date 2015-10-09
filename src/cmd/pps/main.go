package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"

	"go.pedge.io/env"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/proto/client"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog/logrus"

	"google.golang.org/grpc"

	"github.com/spf13/cobra"
	"go.pachyderm.com/pachyderm"
	"go.pachyderm.com/pachyderm/src/pps"
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
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	logrus.Register()
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

	inspectCmd := &cobra.Command{
		Use:  "inspect github.com/user/repository [path/to/specDir]",
		Long: "Inspect a pipeline specification.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 1, Max: 2}, func(args []string) error {
			githubPipelineSource, err := getGithubPipelineSource(args)
			if err != nil {
				return err
			}
			pipelineSource, err := apiClient.CreatePipelineSource(
				context.Background(),
				&pps.CreatePipelineSourceRequest{
					PipelineSource: &pps.PipelineSource{
						TypedPipelineSource: &pps.PipelineSource_GithubPipelineSource{
							GithubPipelineSource: githubPipelineSource,
						},
					},
				},
			)
			if err != nil {
				return err
			}
			pipeline, err := apiClient.CreateAndGetPipeline(
				context.Background(),
				&pps.CreateAndGetPipelineRequest{
					PipelineSourceId: pipelineSource.Id,
				},
			)
			if err != nil {
				return err
			}
			if protoFlag {
				fmt.Printf("%v\n", pipeline)
			} else {
				data, err := json.MarshalIndent(pipeline, "", "\t ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			}
			return nil
		}),
	}
	inspectCmd.Flags().BoolVar(&protoFlag, "proto", false, "Print in proto format instead of JSON.")

	startCmd := &cobra.Command{
		Use:  "start github.com/user/repository [path/to/specDir]",
		Long: "Start a pipeline specification run.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 1, Max: 2}, func(args []string) error {
			githubPipelineSource, err := getGithubPipelineSource(args)
			if err != nil {
				return err
			}
			pipelineSource, err := apiClient.CreatePipelineSource(
				context.Background(),
				&pps.CreatePipelineSourceRequest{
					PipelineSource: &pps.PipelineSource{
						TypedPipelineSource: &pps.PipelineSource_GithubPipelineSource{
							GithubPipelineSource: githubPipelineSource,
						},
					},
				},
			)
			if err != nil {
				return err
			}
			pipeline, err := apiClient.CreateAndGetPipeline(
				context.Background(),
				&pps.CreateAndGetPipelineRequest{
					PipelineSourceId: pipelineSource.Id,
				},
			)
			if err != nil {
				return err
			}
			pipelineRun, err := apiClient.CreatePipelineRun(
				context.Background(),
				&pps.CreatePipelineRunRequest{
					PipelineId: pipeline.Id,
				},
			)
			if err != nil {
				return err
			}
			_, err = apiClient.StartPipelineRun(
				context.Background(),
				&pps.StartPipelineRunRequest{
					PipelineRunId: pipelineRun.Id,
				},
			)
			if err != nil {
				return err
			}
			fmt.Println(pipelineRun.Id)
			return nil
		}),
	}

	statusCmd := &cobra.Command{
		Use:  "status pipelineRunID",
		Long: "Get the status of a pipeline run.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			pipelineRunStatuses, err := apiClient.GetPipelineRunStatus(
				context.Background(),
				&pps.GetPipelineRunStatusRequest{
					PipelineRunId: args[0],
				},
			)
			if err != nil {
				return err
			}
			pipelineRunStatus := pipelineRunStatuses.PipelineRunStatus[0]
			name, ok := pps.PipelineRunStatusType_name[int32(pipelineRunStatus.PipelineRunStatusType)]
			if !ok {
				return fmt.Errorf("unknown run status")
			}
			fmt.Printf("%s %s\n", args[0], strings.Replace(name, "PIPELINE_RUN_STATUS_TYPE_", "", -1))
			return nil
		}),
	}

	logsCmd := &cobra.Command{
		Use:  "logs pipelineRunID [node]",
		Long: "Get the logs from a pipeline run.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 1, Max: 2}, func(args []string) error {
			node := ""
			if len(args) == 2 {
				node = args[1]
			}
			pipelineRunLogs, err := apiClient.GetPipelineRunLogs(
				context.Background(),
				&pps.GetPipelineRunLogsRequest{
					PipelineRunId: args[0],
					Node:          node,
				},
			)
			if err != nil {
				return err
			}
			for _, pipelineRunLog := range pipelineRunLogs.PipelineRunLog {
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
		}),
	}

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

func getGithubPipelineSource(args []string) (*pps.GithubPipelineSource, error) {
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
	return &pps.GithubPipelineSource{
		ContextDir: contextDir,
		User:       split[1],
		Repository: split[2],
	}, nil
}

type logInfo struct {
	Node         string
	ContainerID  string
	Time         time.Time
	OutputStream string
}
