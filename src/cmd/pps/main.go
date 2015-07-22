package main

import (
	"fmt"
	"os"
	"runtime"

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

	rootCmd := &cobra.Command{
		Use: "pps",
		Long: `Access the PPS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PPS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:651.`,
	}
	rootCmd.AddCommand(versionCmd)
	check(rootCmd.Execute())

	os.Exit(0)
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
