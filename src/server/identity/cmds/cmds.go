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
	addConnector := &cobra.Command{
		Short: "Create a new identity provider connector.",
		Long:  `Create a new identity provider connector.`,
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
					ConfigVersion: 0,
					JsonConfig:    string(rawConfigBytes),
				}}

			_, err = c.CreateConnector(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	addConnector.PersistentFlags().StringVar(&id, "id", "", `The id for the new connector.`)
	addConnector.PersistentFlags().StringVar(&name, "name", "", `The user-visible name of the connector.`)
	addConnector.PersistentFlags().StringVar(&t, "type", "", `The type of the connector, ex. github, ldap, saml.`)
	addConnector.PersistentFlags().StringVar(&file, "config", "", `The file to read the JSON-encoded connector configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(addConnector, "idp create connector")
}

// UpdateConnectorCmd returns a cobra.Command to create a new IDP integration
func UpdateConnectorCmd() *cobra.Command {
	var name, file string
	var version int
	updateConnector := &cobra.Command{
		Use:   "{{alias}} <connector id>",
		Short: "Update an existing identity provider connector.",
		Long:  `Update an existing identity provider connector. Only fields which are specified are updated.`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
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
				rawConfigBytes = nil
			}

			req := &identity.UpdateConnectorRequest{
				Config: &identity.ConnectorConfig{
					Id:            args[0],
					Name:          name,
					ConfigVersion: int64(version),
					JsonConfig:    string(rawConfigBytes),
				}}

			_, err = c.UpdateConnector(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	updateConnector.PersistentFlags().StringVar(&name, "name", "", `The new user-visible name of the connector.`)
	updateConnector.PersistentFlags().IntVar(&version, "version", 0, `The new configuration version. This must be one greater than the current version.`)
	updateConnector.PersistentFlags().StringVar(&file, "config", "", `The file to read the JSON-encoded connector configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(updateConnector, "idp update connector")
}

// GetConnectorCmd returns a cobra.Command to get an IDP connector configuration
func GetConnectorCmd() *cobra.Command {
	getConnector := &cobra.Command{
		Use:   "{{alias}} <connector id>",
		Short: "Get the config for an identity provider connector.",
		Long:  "Get the config for an identity provider connector.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}

			resp, err := c.GetConnector(c.Ctx(), &identity.GetConnectorRequest{Id: args[1]})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Printf("Name: %v\nType: %v\nVersion: %v\nConfig:\n%v\n", resp.Config.Name, resp.Config.Type, resp.Config.ConfigVersion, resp.Config.JsonConfig)
			return nil
		}),
	}
	return cmdutil.CreateAlias(getConnector, "idp get connector")
}

// DeleteConnectorCmd returns a cobra.Command to delete an IDP connector
func DeleteConnectorCmd() *cobra.Command {
	deleteConnector := &cobra.Command{
		Short: "Delete an identity provider connector",
		Long:  "Delete an identity provider connector",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}

			_, err = c.DeleteConnector(c.Ctx(), &identity.DeleteConnectorRequest{Id: args[1]})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(deleteConnector, "idp delete connector")
}

// ListConnectorsCmd returns a cobra.Command to list IDP integrations
func ListConnectorsCmd() *cobra.Command {
	listConnectors := &cobra.Command{
		Short: "List identity provider connectors",
		Long:  `List identity provider connectors`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			resp, err := c.ListConnectors(c.Ctx(), &identity.ListConnectorsRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			for _, conn := range resp.Connectors {
				fmt.Printf("%v - %v (%v)", conn.Id, conn.Name, conn.Type)
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(listConnectors, "idp list connector")
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
	return cmdutil.CreateAlias(addConnector, "idp create client")
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
	commands = append(commands, GetConnectorCmd())
	commands = append(commands, UpdateConnectorCmd())
	commands = append(commands, DeleteConnectorCmd())
	commands = append(commands, ListConnectorsCmd())
	commands = append(commands, CreateClientCmd())

	return commands
}
