package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/ppsutil"
	"github.com/peter-edge/go-env"
	"github.com/spf13/cobra"
)

const (
	defaultAddress = "0.0.0.0:651"
)

type appEnv struct {
	Address string `env:"PPS_ADDRESS"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	appEnv := &appEnv{}
	check(env.Populate(appEnv, env.PopulateOptions{}))
	if appEnv.Address == "" {
		appEnv.Address = defaultAddress
	}

	clientConn, err := grpc.Dial(appEnv.Address)
	check(err)
	apiClient := pps.NewApiClient(clientConn)

	versionCmd := &cobra.Command{
		Use:  "version",
		Long: "Print the version.",
		Run: func(cmd *cobra.Command, args []string) {
			getVersionResponse, err := ppsutil.GetVersion(apiClient)
			check(err)
			fmt.Printf("Client: %s\nServer: %s\n", common.VersionString(), pps.VersionString(getVersionResponse.Version))
		},
	}

	var protoFlag bool
	inspectCmd := &cobra.Command{
		Use:  "inspect",
		Long: "Inspect a pipeline specification.",
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			if !strings.HasPrefix(path, "github.com/") {
				check(fmt.Errorf("%s is not supported", path))
			}
			split := strings.Split(path, "/")
			if len(split) != 3 {
				check(fmt.Errorf("%s is not supported", path))
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
			check(err)
			if protoFlag {
				fmt.Printf("%v\n", getPipelineResponse.Pipeline)
			} else {
				data, err := json.MarshalIndent(getPipelineResponse.Pipeline, "", "\t ")
				check(err)
				fmt.Println(string(data))
			}
		},
	}
	inspectCmd.Flags().BoolVar(&protoFlag, "proto", false, "Print in proto format instead of JSON.")

	rootCmd := &cobra.Command{
		Use: "pps",
		Long: `Access the PPS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PPS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:651.`,
	}
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(inspectCmd)
	check(rootCmd.Execute())

	os.Exit(0)
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
