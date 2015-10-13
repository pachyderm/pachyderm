package main

import (
	"fmt"
	"strings"

	"go.pedge.io/env"
	"go.pedge.io/protolog/logrus"

	"google.golang.org/grpc"

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
	fmt.Println(clientConn)
	rootCmd := &cobra.Command{
		Use: "pps",
		Long: `Access the PPS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PPS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:651.`,
	}
	return rootCmd.Execute()
}
