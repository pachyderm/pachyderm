package cmds

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
)

// GetConfigCmd returns a cobra command that lets the caller see the configured
// auth backends in Pachyderm
func GetConfigCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	var enterprise bool
	var format string
	getConfig := &cobra.Command{
		Short: "Retrieve Pachyderm's current auth configuration",
		Long:  "Retrieve Pachyderm's current auth configuration",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), enterprise)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer errors.Close(&retErr, c, "close client")
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
			e, err := serde.GetEncoder(format, &buf, serde.WithIndent(2),
				serde.WithOrigName(true))
			if err != nil {
				return err
			}
			// Use Encode() rather than EncodeProto, because the official proto->json
			// spec (https://developers.google.com/protocol-buffers/docs/proto3#json)
			// requires that int64 fields (e.g. live_config_version) be serialized as
			// strings rather than ints, which would break existing auth configs. Go's
			// built-in json serializer marshals int64 fields to JSON numbers
			if err := e.Encode(resp.Configuration); err != nil {
				return errors.EnsureStack(err)
			}
			fmt.Println(buf.String())
			return nil
		}),
	}
	getConfig.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "Get auth config for the active enterprise context")
	getConfig.Flags().StringVarP(&format, "output-format", "o", "json", "output "+
		"format (\"json\" or \"yaml\")")
	return cmdutil.CreateAlias(getConfig, "auth get-config")
}

// SetConfigCmd returns a cobra command that lets the caller configure auth
// backends in Pachyderm
func SetConfigCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	var enterprise bool
	var file string
	setConfig := &cobra.Command{
		Short: "Set Pachyderm's current auth configuration",
		Long:  "Set Pachyderm's current auth configuration",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), enterprise)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer errors.Close(&retErr, c, "close client")
			var rawConfigBytes []byte
			if file == "-" {
				var err error
				rawConfigBytes, err = io.ReadAll(os.Stdin)
				if err != nil {
					return errors.Wrapf(err, "could not read config from stdin")
				}
			} else if file != "" {
				var err error
				rawConfigBytes, err = os.ReadFile(file)
				if err != nil {
					return errors.Wrapf(err, "could not read config from %q", file)
				}
			} else {
				return errors.New("must set input file (use \"-\" to read from stdin)")
			}

			// parse config
			var config auth.OIDCConfig
			if err := serde.Decode(rawConfigBytes, &config); err != nil {
				return errors.Wrapf(err, "could not parse config")
			}
			// TODO(msteffen): try to handle empty config?
			_, err = c.SetConfiguration(c.Ctx(), &auth.SetConfigurationRequest{
				Configuration: &config,
			})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	setConfig.PersistentFlags().BoolVar(&enterprise, "enterprise", false, "Set auth config for the active enterprise context")
	setConfig.Flags().StringVarP(&file, "file", "f", "-", "input file (to use "+
		"as the new config")
	return cmdutil.CreateAlias(setConfig, "auth set-config")
}
