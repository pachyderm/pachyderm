package cmds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

// GetConfig returns a cobra command that lets the caller see the configured
// auth backends in Pachyderm
func GetConfig() *cobra.Command {
	var format string
	getConfig := &cobra.Command{
		Use:   "get-config",
		Short: "Retrieve Pachyderm's current auth configuration",
		Long:  "Retrieve Pachyderm's current auth configuration",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %v", err)
			}
			defer c.Close()
			resp, err := c.GetConfiguration(c.Ctx(), &auth.GetConfigurationRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if resp.Configuration == nil {
				fmt.Println("no auth config set")
				return nil
			}
			output, err := json.MarshalIndent(resp.Configuration, "", "  ")
			if err != nil {
				return fmt.Errorf("could not marshal response:\n%v\ndue to: %v", resp.Configuration, err)
			}
			switch format {
			case "json":
				// already done
			case "yaml":
				output, err = yaml.JSONToYAML(output)
				if err != nil {
					return fmt.Errorf("could not convert json to yaml: %v", err)
				}
			default:
				return fmt.Errorf("invalid output format: %v", format)
			}
			fmt.Println(string(output))
			return nil
		}),
	}
	getConfig.Flags().StringVarP(&format, "output-format", "o", "json", "output "+
		"format (\"json\" or \"yaml\")")
	return getConfig
}

// SetConfig returns a cobra command that lets the caller configure auth
// backends in Pachyderm
func SetConfig() *cobra.Command {
	var file string
	configure := &cobra.Command{
		Use:   "set-config",
		Short: "Set Pachyderm's current auth configuration",
		Long:  "Set Pachyderm's current auth configuration",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine(true, "user")
			if err != nil {
				return fmt.Errorf("could not connect: %v", err)
			}
			defer c.Close()
			var configBytes []byte
			if file == "-" {
				var err error
				configBytes, err = ioutil.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("could not read config from stdin: %v", err)
				}
			} else if file != "" {
				var err error
				configBytes, err = ioutil.ReadFile(file)
				if err != nil {
					return fmt.Errorf("could not read config from %q: %v", file, err)
				}
			} else {
				return errors.New("must set input file (use \"-\" to read from stdin)")
			}

			// Try to parse config as YAML (JSON is a subset of YAML)
			var config auth.AuthConfig
			if err := yaml.Unmarshal(configBytes, &config); err != nil {
				return fmt.Errorf("could not parse config: %v", err)
			}
			// TODO(msteffen): try to handle empty config?
			_, err = c.SetConfiguration(c.Ctx(), &auth.SetConfigurationRequest{
				Configuration: &config,
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	configure.Flags().StringVarP(&file, "file", "f", "-", "input file (to use "+
		"as the new config")
	return configure
}
