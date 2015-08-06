package main

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/pkg/cobramainutil"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/client"
	"k8s.io/kubernetes/pkg/client/clientcmd"
)

var (
	defaultEnv = map[string]string{}
)

type appEnv struct{}

func main() {
	mainutil.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)

	runCmd := cobramainutil.Command{
		Use:  "run",
		Long: "run",
		Run: func(cmd *cobra.Command, args []string) error {
			return run(appEnv, args)
		},
	}.ToCobraCommand()

	rootCmd := &cobra.Command{
		Use:  "pachkube",
		Long: `pachkube`,
	}

	rootCmd.AddCommand(runCmd)
	return rootCmd.Execute()
}

func run(appEnv *appEnv, args []string) error {
	clientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return err
	}
	client, err := client.New(clientConfig)
	if err != nil {
		return err
	}
	versionInfo, err := client.ServerVersion()
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", versionInfo)
	return nil
}
