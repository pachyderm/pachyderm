package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/serde"

	"github.com/spf13/cobra"
)

// GetConfigCmd returns a cobra command that lets the caller see the configured
// auth backends in Pachyderm
func GetConfigCmd() *cobra.Command {
	var format string
	getConfig := &cobra.Command{
		Short: "Retrieve Pachyderm's current auth configuration",
		Long:  "Retrieve Pachyderm's current auth configuration",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
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

			var buf bytes.Buffer
			if format == "" {
				format = "json"
			} else {
				format = strings.ToLower(format)
			}
			e, err := serde.GetEncoder(format, &buf)
			if err != nil {
				return err
			}
			if err := e.EncodeProto(resp.Configuration); err != nil {
				return err
			}
			fmt.Println(buf.String())
			return nil
		}),
	}
	getConfig.Flags().StringVarP(&format, "output-format", "o", "json", "output "+
		"format (\"json\" or \"yaml\")")
	return cmdutil.CreateAlias(getConfig, "auth get-config")
}

// SetConfigCmd returns a cobra command that lets the caller configure auth
// backends in Pachyderm
func SetConfigCmd() *cobra.Command {
	var file string
	setConfig := &cobra.Command{
		Short: "Set Pachyderm's current auth configuration",
		Long:  "Set Pachyderm's current auth configuration",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return fmt.Errorf("could not connect: %v", err)
			}
			defer c.Close()
			var rawConfigBytes []byte
			if file == "-" {
				var err error
				rawConfigBytes, err = ioutil.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("could not read config from stdin: %v", err)
				}
			} else if file != "" {
				var err error
				rawConfigBytes, err = ioutil.ReadFile(file)
				if err != nil {
					return fmt.Errorf("could not read config from %q: %v", file, err)
				}
			} else {
				return errors.New("must set input file (use \"-\" to read from stdin)")
			}

			// parse config
			var config auth.AuthConfig
			if err := serde.DecodeYAML(rawConfigBytes, &config); err != nil {
				return fmt.Errorf("could not parse config: %v", err)
			}
			// TODO(msteffen): try to handle empty config?
			_, err = c.SetConfiguration(c.Ctx(), &auth.SetConfigurationRequest{
				Configuration: &config,
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	setConfig.Flags().StringVarP(&file, "file", "f", "-", "input file (to use "+
		"as the new config")
	return cmdutil.CreateAlias(setConfig, "auth set-config")
}
