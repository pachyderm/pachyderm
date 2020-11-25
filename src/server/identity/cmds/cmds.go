package cmds

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/cobra"
)

// CreateConnectorCmd returns a cobra.Command to create a new IDP integration
func CreateConnectorCmd() *cobra.Command {
	var id, name, t, file string
	var version int
	addConnector := &cobra.Command{
		Short: "Create a new identity provider connector",
		Long:  `Create a new identity provider connector`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			var rawConfigBytes []byte
			if file == "-" {
				var err error
				rawConfigBytes, err = ioutil.ReadAll(os.Stdin)
				if err != nil {
					return errors.Wrapf(err, "could not read config from stdin")
				}
			} else if file != "" {
				var err error
				rawConfigBytes, err = ioutil.ReadFile(file)
				if err != nil {
					return errors.Wrapf(err, "could not read config from %q", file)
				}
			} else {
				return errors.New("must set input file (use \"-\" to read from stdin)")
			}

			req := &identity.CreateConnectorRequest{
				Config: &identity.ConnectorConfig{
					Id:            id,
					Name:          name,
					Type:          t,
					ConfigVersion: int64(version),
					JsonConfig:    string(rawConfigBytes),
				}}

			_, err = c.CreateConnector(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	addConnector.PersistentFlags().StringVar(&id, "id", "", ``)
	addConnector.PersistentFlags().StringVar(&name, "name", "", ``)
	addConnector.PersistentFlags().StringVar(&t, "type", "", ``)
	addConnector.PersistentFlags().IntVar(&version, "version", 0, ``)
	addConnector.PersistentFlags().StringVar(&file, "config", "", ``)
	return cmdutil.CreateAlias(addConnector, "idp connector create")
}

// ListConnectorsCmd returns a cobra.Command to list IDP integrations
func ListConnectorsCmd() *cobra.Command {
	var id, name, t, file string
	var version int
	addConnector := &cobra.Command{
		Short: "Create a new identity provider connector",
		Long:  `Create a new identity provider connector`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()
			var rawConfigBytes []byte
			if file == "-" {
				var err error
				rawConfigBytes, err = ioutil.ReadAll(os.Stdin)
				if err != nil {
					return errors.Wrapf(err, "could not read config from stdin")
				}
			} else if file != "" {
				var err error
				rawConfigBytes, err = ioutil.ReadFile(file)
				if err != nil {
					return errors.Wrapf(err, "could not read config from %q", file)
				}
			} else {
				return errors.New("must set input file (use \"-\" to read from stdin)")
			}

			req := &identity.CreateConnectorRequest{
				Config: &identity.ConnectorConfig{
					Id:            id,
					Name:          name,
					Type:          t,
					ConfigVersion: int64(version),
					JsonConfig:    string(rawConfigBytes),
				}}

			_, err = c.CreateConnector(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	addConnector.PersistentFlags().StringVar(&id, "id", "", ``)
	addConnector.PersistentFlags().StringVar(&name, "name", "", ``)
	addConnector.PersistentFlags().StringVar(&t, "type", "", ``)
	addConnector.PersistentFlags().IntVar(&version, "version", 0, ``)
	addConnector.PersistentFlags().StringVar(&file, "config", "", ``)
	return cmdutil.CreateAlias(addConnector, "idp connector create")
}

// CreateClientCmd returns a cobra.Command to create a new OIDC client
func CreateClientCmd() *cobra.Command {
	var id, name, redirectURI string
	addConnector := &cobra.Command{
		Short: "Create a new OIDC client",
		Long:  `Create a new OIDC client`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			req := &identity.CreateClientRequest{
				Client: &identity.Client{
					Id:           id,
					Name:         name,
					RedirectUris: []string{redirectURI},
				},
			}

			resp, err := c.CreateClient(c.Ctx(), req)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			fmt.Printf("Client secret: %q\n", resp.Client.Secret)
			return nil
		}),
	}
	addConnector.PersistentFlags().StringVar(&id, "id", "", ``)
	addConnector.PersistentFlags().StringVar(&name, "name", "", ``)
	addConnector.PersistentFlags().StringVar(&redirectURI, "redirectUri", "", ``)
	return cmdutil.CreateAlias(addConnector, "idp client create")
}

// Cmds returns a list of cobra commands for authenticating and authorizing
// users in an auth-enabled Pachyderm cluster.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	idp := &cobra.Command{
		Short: "Commands to manage identity provider integrations",
		Long:  "Commands to manage identity provider integrations",
	}

	commands = append(commands, cmdutil.CreateAlias(idp, "idp"))
	commands = append(commands, CreateConnectorCmd())
	commands = append(commands, CreateClientCmd())

	return commands
}
