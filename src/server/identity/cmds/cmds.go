package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/cobra"
)

// SetIdentityConfigCmd returns a cobra.Command to configure the identity server
func SetIdentityConfigCmd() *cobra.Command {
	var issuer string
	setConfig := &cobra.Command{
		Short: "Set the identity server config",
		Long:  `Set the identity server config`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			req := &identity.SetIdentityConfigRequest{
				Config: &identity.IdentityConfig{
					Issuer: issuer,
				}}

			_, err = c.SetIdentityConfig(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	setConfig.PersistentFlags().StringVar(&issuer, "issuer", "", `The issuer for the identity server.`)
	return cmdutil.CreateAlias(setConfig, "idp set config")
}

// GetIdentityConfigCmd returns a cobra.Command to fetch the current ID server config
func GetIdentityConfigCmd() *cobra.Command {
	getConfig := &cobra.Command{
		Short: "Get the identity server config",
		Long:  `Get the identity server config`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			resp, err := c.GetIdentityConfig(c.Ctx(), &identity.GetIdentityConfigRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Printf("Issuer: %q\n", resp.Config.Issuer)
			return nil
		}),
	}
	return cmdutil.CreateAlias(getConfig, "idp get config")
}

// CreateIDPConnectorCmd returns a cobra.Command to create a new IDP integration
func CreateIDPConnectorCmd() *cobra.Command {
	var id, name, t, file string
	createConnector := &cobra.Command{
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

			req := &identity.CreateIDPConnectorRequest{
				Config: &identity.IDPConnector{
					Id:            id,
					Name:          name,
					Type:          t,
					ConfigVersion: 0,
					JsonConfig:    string(rawConfigBytes),
				}}

			_, err = c.CreateIDPConnector(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	createConnector.PersistentFlags().StringVar(&id, "id", "", `The id for the new connector.`)
	createConnector.PersistentFlags().StringVar(&name, "name", "", `The user-visible name of the connector.`)
	createConnector.PersistentFlags().StringVar(&t, "type", "", `The type of the connector, ex. github, ldap, saml.`)
	createConnector.PersistentFlags().StringVar(&file, "config", "", `The file to read the JSON-encoded connector configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(createConnector, "idp create connector")
}

// UpdateIDPConnectorCmd returns a cobra.Command to create a new IDP integration
func UpdateIDPConnectorCmd() *cobra.Command {
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

			req := &identity.UpdateIDPConnectorRequest{
				Config: &identity.IDPConnector{
					Id:            args[0],
					Name:          name,
					ConfigVersion: int64(version),
					JsonConfig:    string(rawConfigBytes),
				}}

			_, err = c.UpdateIDPConnector(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	updateConnector.PersistentFlags().StringVar(&name, "name", "", `The new user-visible name of the connector.`)
	updateConnector.PersistentFlags().IntVar(&version, "version", 0, `The new configuration version. This must be one greater than the current version.`)
	updateConnector.PersistentFlags().StringVar(&file, "config", "", `The file to read the JSON-encoded connector configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(updateConnector, "idp update connector")
}

// GetIDPConnectorCmd returns a cobra.Command to get an IDP connector configuration
func GetIDPConnectorCmd() *cobra.Command {
	getConnector := &cobra.Command{
		Use:   "{{alias}} <connector id>",
		Short: "Get the config for an identity provider connector.",
		Long:  "Get the config for an identity provider connector.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}

			resp, err := c.GetIDPConnector(c.Ctx(), &identity.GetIDPConnectorRequest{Id: args[1]})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Printf("Name: %v\nType: %v\nVersion: %v\nConfig:\n%v\n", resp.Config.Name, resp.Config.Type, resp.Config.ConfigVersion, resp.Config.JsonConfig)
			return nil
		}),
	}
	return cmdutil.CreateAlias(getConnector, "idp get connector")
}

// DeleteIDPConnectorCmd returns a cobra.Command to delete an IDP connector
func DeleteIDPConnectorCmd() *cobra.Command {
	deleteConnector := &cobra.Command{
		Short: "Delete an identity provider connector",
		Long:  "Delete an identity provider connector",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}

			_, err = c.DeleteIDPConnector(c.Ctx(), &identity.DeleteIDPConnectorRequest{Id: args[1]})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(deleteConnector, "idp delete connector")
}

// ListIDPConnectorsCmd returns a cobra.Command to list IDP integrations
func ListIDPConnectorsCmd() *cobra.Command {
	listConnectors := &cobra.Command{
		Short: "List identity provider connectors",
		Long:  `List identity provider connectors`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			resp, err := c.ListIDPConnectors(c.Ctx(), &identity.ListIDPConnectorsRequest{})
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

// CreateOIDCClientCmd returns a cobra.Command to create a new OIDC client
func CreateOIDCClientCmd() *cobra.Command {
	var id, name, redirectURI, trustedPeers string
	createClient := &cobra.Command{
		Short: "Create a new OIDC client.",
		Long:  `Create a new OIDC client.`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			req := &identity.CreateOIDCClientRequest{
				Client: &identity.OIDCClient{
					Id:           id,
					Name:         name,
					RedirectUris: strings.Split(redirectURI, ","),
					TrustedPeers: strings.Split(trustedPeers, ","),
				},
			}

			resp, err := c.CreateOIDCClient(c.Ctx(), req)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			fmt.Printf("Client secret: %q\n", resp.Client.Secret)
			return nil
		}),
	}
	createClient.PersistentFlags().StringVar(&id, "id", "", `The client_id of the new client.`)
	createClient.PersistentFlags().StringVar(&name, "name", "", `The user-visible name of the new client.`)
	createClient.PersistentFlags().StringVar(&redirectURI, "redirectUris", "", `A comma-separated list of authorized redirect URLs for callbacks.`)
	createClient.PersistentFlags().StringVar(&trustedPeers, "trustedPeers", "", `A comma-separated list of clients who can get tokens for this service.`)
	return cmdutil.CreateAlias(createClient, "idp create client")
}

// DeleteOIDCClientCmd returns a cobra.Command to delete an OIDC client
func DeleteOIDCClientCmd() *cobra.Command {
	deleteClient := &cobra.Command{
		Use:   "{{alias}} <client ID>",
		Short: "Delete an OIDC client.",
		Long:  `Delete an OIDC client.`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			_, err = c.DeleteOIDCClient(c.Ctx(), &identity.DeleteOIDCClientRequest{Id: args[1]})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(deleteClient, "idp delete client")
}

// GetOIDCClientCmd returns a cobra.Command to get an OIDC client
func GetOIDCClientCmd() *cobra.Command {
	getClient := &cobra.Command{
		Use:   "{{alias}} <client ID>",
		Short: "Get an OIDC client.",
		Long:  `Get an OIDC client.`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			resp, err := c.GetOIDCClient(c.Ctx(), &identity.GetOIDCClientRequest{Id: args[1]})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			fmt.Printf("client_id: %v\nsecret: %v\nname: %v\nredirect URIs: %v\ntrusted peers: %v\n", resp.Client.Id, resp.Client.Secret, resp.Client.Name, strings.Join(resp.Client.RedirectUris, ", "), strings.Join(resp.Client.TrustedPeers, ", "))
			return nil
		}),
	}
	return cmdutil.CreateAlias(getClient, "idp get client")
}

// UpdateOIDCClientCmd returns a cobra.Command to update an existing OIDC client
func UpdateOIDCClientCmd() *cobra.Command {
	var name, redirectURI, trustedPeers string
	updateClient := &cobra.Command{
		Use:   "{{ alias }} <client ID>",
		Short: "Update an OIDC client.",
		Long:  `Update an OIDC client.`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			req := &identity.UpdateOIDCClientRequest{
				Client: &identity.OIDCClient{
					Id:           args[0],
					Name:         name,
					RedirectUris: strings.Split(redirectURI, ","),
					TrustedPeers: strings.Split(trustedPeers, ","),
				},
			}

			_, err = c.UpdateOIDCClient(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	updateClient.PersistentFlags().StringVar(&name, "name", "", `The user-visible name of the new client.`)
	updateClient.PersistentFlags().StringVar(&redirectURI, "redirectUris", "", `A comma-separated list of redirect URLs for callbacks.`)
	updateClient.PersistentFlags().StringVar(&trustedPeers, "trustedPeers", "", `A comma-separated list of clients that can get tokens for this service`)
	return cmdutil.CreateAlias(updateClient, "idp update client")
}

// ListOIDCClientsCmd returns a cobra.Command to list IDP integrations
func ListOIDCClientsCmd() *cobra.Command {
	listConnectors := &cobra.Command{
		Short: "List OIDC clients.",
		Long:  `List OIDC clients.`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			resp, err := c.ListOIDCClients(c.Ctx(), &identity.ListOIDCClientsRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			for _, client := range resp.Clients {
				fmt.Printf("%v", client.Id)
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(listConnectors, "idp list client")
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
	commands = append(commands, GetIdentityConfigCmd())
	commands = append(commands, SetIdentityConfigCmd())
	commands = append(commands, CreateIDPConnectorCmd())
	commands = append(commands, GetIDPConnectorCmd())
	commands = append(commands, UpdateIDPConnectorCmd())
	commands = append(commands, DeleteIDPConnectorCmd())
	commands = append(commands, ListIDPConnectorsCmd())
	commands = append(commands, CreateOIDCClientCmd())
	commands = append(commands, GetOIDCClientCmd())
	commands = append(commands, UpdateOIDCClientCmd())
	commands = append(commands, ListOIDCClientsCmd())

	return commands
}
