/*
Package main contains a CLI binary for the dockervolume API functionality not contained
within the docker volume plugin API.
*/
package main

import (
	"fmt"
	"os"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"go.pedge.io/dockervolume"
	"go.pedge.io/env"
	"google.golang.org/grpc"
)

var (
	defaultAddress = fmt.Sprintf("0.0.0.0:%d", dockervolume.DefaultGRPCPort)
	defaultEnv     = map[string]string{
		"ADDRESS": defaultAddress,
	}
	marshaler = &jsonpb.Marshaler{
		Indent: "  ",
	}
)

type appEnv struct {
	Address string `env:"ADDRESS"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)

	cleanup := &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup all existing volumes.",
		Long:  "Cleanup all existing volumes that the volume driver is currently handling.",
		Run: cobraFunc(0, func(_ []string) error {
			client, err := getClient(appEnv)
			if err != nil {
				return err
			}
			response, err := client.Cleanup()
			if err != nil {
				return err
			}
			for _, element := range response {
				if err := marshal(element); err != nil {
					return err
				}
			}
			return nil
		}),
	}

	getVolume := &cobra.Command{
		Use:   "get-volume name",
		Short: "Get a volume by name.",
		Long:  "Get a volume by name.",
		Run: cobraFunc(1, func(args []string) error {
			client, err := getClient(appEnv)
			if err != nil {
				return err
			}
			response, err := client.GetVolume(args[0])
			if err != nil {
				return err
			}
			return marshal(response)
		}),
	}

	listVolumes := &cobra.Command{
		Use:   "list-volumes",
		Short: "List all volumes controlled by this driver",
		Long:  "List all volumes controlled by this driver",
		Run: cobraFunc(0, func(_ []string) error {
			client, err := getClient(appEnv)
			if err != nil {
				return err
			}
			response, err := client.ListVolumes()
			if err != nil {
				return err
			}
			for _, element := range response {
				if err := marshal(element); err != nil {
					return err
				}
			}
			return nil
		}),
	}

	rootCmd := &cobra.Command{
		Use:   "dockervolume",
		Short: "Access a Docker volume driver.",
		Long:  fmt.Sprintf("Access a Dockervolume driver.\n\nThe environment variable ADDRESS controls what server the CLI connects to, the default is %s.", defaultAddress),
	}
	rootCmd.AddCommand(cleanup)
	rootCmd.AddCommand(getVolume)
	rootCmd.AddCommand(listVolumes)
	return rootCmd.Execute()
}

func cobraFunc(expectedNumArgs int, f func([]string) error) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, args []string) {
		check(checkArgs(args, expectedNumArgs))
		check(f(args))
	}
}

func checkArgs(args []string, expected int) error {
	if len(args) != expected {
		return fmt.Errorf("Wrong number of arguments. Got %d, need %d.", len(args), expected)
	}
	return nil
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func getClient(appEnv *appEnv) (dockervolume.VolumeDriverClient, error) {
	clientConn, err := grpc.Dial(appEnv.Address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return dockervolume.NewVolumeDriverClient(dockervolume.NewAPIClient(clientConn)), nil
}

func marshal(message proto.Message) error {
	if err := marshaler.Marshal(os.Stdout, message); err != nil {
		return err
	}
	fmt.Println()
	return nil
}
