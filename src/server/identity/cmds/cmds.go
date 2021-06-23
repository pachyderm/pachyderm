package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"

	"github.com/spf13/cobra"
)

type connectorConfig struct {
	ID      string
	Name    string
	Type    string
	Version int64
	Config  interface{}
}

func newConnectorConfig(conn *identity.IDPConnector) (*connectorConfig, error) {
	config := connectorConfig{
		ID:      conn.Id,
		Name:    conn.Name,
		Type:    conn.Type,
		Version: conn.ConfigVersion,
	}

	if err := json.Unmarshal([]byte(conn.JsonConfig), &config.Config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c connectorConfig) toIDPConnector() (*identity.IDPConnector, error) {
	jsonConfig, err := json.Marshal(c.Config)
	if err != nil {
		return nil, err
	}
	return &identity.IDPConnector{
		Id:            c.ID,
		Name:          c.Name,
		Type:          c.Type,
		ConfigVersion: c.Version,
		JsonConfig:    string(jsonConfig),
	}, nil
}

func deserializeYAML(env cmdutil.Env, file string, target interface{}) error {
	var rawConfigBytes []byte
	if file == "-" {
		var err error
		rawConfigBytes, err = ioutil.ReadAll(env.In())
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

	return serde.DecodeYAML(rawConfigBytes, target)
}

// SetIdentityServerConfigCmd returns a cobra.Command to configure the identity server
func SetIdentityServerConfigCmd() *cobra.Command {
	var file string
	setConfig := &cobra.Command{
		Short: "Set the identity server config",
		Long:  `Set the identity server config`,
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			var config identity.IdentityServerConfig
			if err := deserializeYAML(env, file, &config); err != nil {
				return errors.Wrapf(err, "unable to parse config")
			}

			c := env.EnterpriseClient("user")
			_, err := c.SetIdentityServerConfig(c.Ctx(), &identity.SetIdentityServerConfigRequest{Config: &config})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	setConfig.PersistentFlags().StringVar(&file, "config", "-", `The file to read the YAML-encoded configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(setConfig, "idp set-config")
}

// GetIdentityServerConfigCmd returns a cobra.Command to fetch the current ID server config
func GetIdentityServerConfigCmd() *cobra.Command {
	getConfig := &cobra.Command{
		Short: "Get the identity server config",
		Long:  `Get the identity server config`,
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			resp, err := c.GetIdentityServerConfig(c.Ctx(), &identity.GetIdentityServerConfigRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			yamlStr, err := serde.EncodeYAML(resp.Config)
			if err != nil {
				return err
			}
			fmt.Fprintln(env.Out(), string(yamlStr))
			return nil
		}),
	}
	return cmdutil.CreateAlias(getConfig, "idp get-config")
}

// CreateIDPConnectorCmd returns a cobra.Command to create a new IDP integration
func CreateIDPConnectorCmd() *cobra.Command {
	var file string
	createConnector := &cobra.Command{
		Short: "Create a new identity provider connector.",
		Long:  `Create a new identity provider connector.`,
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			var connector connectorConfig
			if err := deserializeYAML(env, file, &connector); err != nil {
				return errors.Wrapf(err, "unable to parse config")
			}

			config, err := connector.toIDPConnector()
			if err != nil {
				return err
			}

			c := env.EnterpriseClient("user")
			_, err = c.CreateIDPConnector(c.Ctx(), &identity.CreateIDPConnectorRequest{Connector: config})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	createConnector.PersistentFlags().StringVar(&file, "config", "-", `The file to read the YAML-encoded connector configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(createConnector, "idp create-connector")
}

// UpdateIDPConnectorCmd returns a cobra.Command to create a new IDP integration
func UpdateIDPConnectorCmd() *cobra.Command {
	var file string
	updateConnector := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Update an existing identity provider connector.",
		Long:  `Update an existing identity provider connector. Only fields which are specified are updated.`,
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			var connector connectorConfig
			if err := deserializeYAML(env, file, &connector); err != nil {
				return errors.Wrapf(err, "unable to parse config")
			}

			config, err := connector.toIDPConnector()
			if err != nil {
				return err
			}

			req := &identity.UpdateIDPConnectorRequest{Connector: config}

			c := env.EnterpriseClient("user")
			_, err = c.UpdateIDPConnector(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	updateConnector.PersistentFlags().StringVar(&file, "config", "-", `The file to read the YAML-encoded connector configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(updateConnector, "idp update-connector")
}

// GetIDPConnectorCmd returns a cobra.Command to get an IDP connector configuration
func GetIDPConnectorCmd() *cobra.Command {
	getConnector := &cobra.Command{
		Use:   "{{alias}} <connector id>",
		Short: "Get the config for an identity provider connector.",
		Long:  "Get the config for an identity provider connector.",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			resp, err := c.GetIDPConnector(c.Ctx(), &identity.GetIDPConnectorRequest{Id: args[0]})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			config, err := newConnectorConfig(resp.Connector)
			if err != nil {
				return err
			}
			yamlStr, err := serde.EncodeYAML(config)
			if err != nil {
				return err
			}
			fmt.Fprintln(env.Out(), string(yamlStr))
			return nil
		}),
	}
	return cmdutil.CreateAlias(getConnector, "idp get-connector")
}

// DeleteIDPConnectorCmd returns a cobra.Command to delete an IDP connector
func DeleteIDPConnectorCmd() *cobra.Command {
	deleteConnector := &cobra.Command{
		Short: "Delete an identity provider connector",
		Long:  "Delete an identity provider connector",
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			_, err := c.DeleteIDPConnector(c.Ctx(), &identity.DeleteIDPConnectorRequest{Id: args[0]})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(deleteConnector, "idp delete-connector")
}

// ListIDPConnectorsCmd returns a cobra.Command to list IDP integrations
func ListIDPConnectorsCmd() *cobra.Command {
	listConnectors := &cobra.Command{
		Short: "List identity provider connectors",
		Long:  `List identity provider connectors`,
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			resp, err := c.ListIDPConnectors(c.Ctx(), &identity.ListIDPConnectorsRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			for _, conn := range resp.Connectors {
				fmt.Fprintf(env.Out(), "%v - %v (%v)\n", conn.Id, conn.Name, conn.Type)
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(listConnectors, "idp list-connector")
}

// CreateOIDCClientCmd returns a cobra.Command to create a new OIDC client
func CreateOIDCClientCmd() *cobra.Command {
	var file string
	createClient := &cobra.Command{
		Short: "Create a new OIDC client.",
		Long:  `Create a new OIDC client.`,
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			var client identity.OIDCClient
			if err := deserializeYAML(env, file, &client); err != nil {
				return errors.Wrapf(err, "unable to parse config")
			}

			c := env.EnterpriseClient("user")
			resp, err := c.CreateOIDCClient(c.Ctx(), &identity.CreateOIDCClientRequest{Client: &client})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			fmt.Fprintf(env.Out(), "Client secret: %q\n", resp.Client.Secret)
			return nil
		}),
	}
	createClient.PersistentFlags().StringVar(&file, "config", "-", `The file to read the YAML-encoded client configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(createClient, "idp create-client")
}

// DeleteOIDCClientCmd returns a cobra.Command to delete an OIDC client
func DeleteOIDCClientCmd() *cobra.Command {
	deleteClient := &cobra.Command{
		Use:   "{{alias}} <client ID>",
		Short: "Delete an OIDC client.",
		Long:  `Delete an OIDC client.`,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			_, err := c.DeleteOIDCClient(c.Ctx(), &identity.DeleteOIDCClientRequest{Id: args[0]})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(deleteClient, "idp delete-client")
}

// GetOIDCClientCmd returns a cobra.Command to get an OIDC client
func GetOIDCClientCmd() *cobra.Command {
	getClient := &cobra.Command{
		Use:   "{{alias}} <client ID>",
		Short: "Get an OIDC client.",
		Long:  `Get an OIDC client.`,
		RunE: cmdutil.RunFixedArgs(1, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			resp, err := c.GetOIDCClient(c.Ctx(), &identity.GetOIDCClientRequest{Id: args[0]})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			yamlStr, err := serde.EncodeYAML(resp.Client)
			if err != nil {
				return err
			}
			fmt.Fprintln(env.Out(), string(yamlStr))
			return nil
		}),
	}
	return cmdutil.CreateAlias(getClient, "idp get-client")
}

// UpdateOIDCClientCmd returns a cobra.Command to update an existing OIDC client
func UpdateOIDCClientCmd() *cobra.Command {
	var file string
	updateClient := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Update an OIDC client.",
		Long:  `Update an OIDC client.`,
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			var client identity.OIDCClient
			if err := deserializeYAML(env, file, &client); err != nil {
				return errors.Wrapf(err, "unable to parse config")
			}

			c := env.EnterpriseClient("user")
			_, err := c.UpdateOIDCClient(c.Ctx(), &identity.UpdateOIDCClientRequest{Client: &client})
			return grpcutil.ScrubGRPC(err)
		}),
	}

	updateClient.PersistentFlags().StringVar(&file, "config", "-", `The file to read the YAML-encoded client configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(updateClient, "idp update-client")
}

// ListOIDCClientsCmd returns a cobra.Command to list IDP integrations
func ListOIDCClientsCmd() *cobra.Command {
	listConnectors := &cobra.Command{
		Short: "List OIDC clients.",
		Long:  `List OIDC clients.`,
		RunE: cmdutil.RunFixedArgs(0, func(args []string, env cmdutil.Env) error {
			c := env.EnterpriseClient("user")
			resp, err := c.ListOIDCClients(c.Ctx(), &identity.ListOIDCClientsRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			for _, client := range resp.Clients {
				fmt.Fprintf(env.Out(), "%v\n", client.Id)
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(listConnectors, "idp list-client")
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
	commands = append(commands, GetIdentityServerConfigCmd())
	commands = append(commands, SetIdentityServerConfigCmd())
	commands = append(commands, CreateIDPConnectorCmd())
	commands = append(commands, GetIDPConnectorCmd())
	commands = append(commands, UpdateIDPConnectorCmd())
	commands = append(commands, DeleteIDPConnectorCmd())
	commands = append(commands, ListIDPConnectorsCmd())
	commands = append(commands, CreateOIDCClientCmd())
	commands = append(commands, GetOIDCClientCmd())
	commands = append(commands, UpdateOIDCClientCmd())
	commands = append(commands, DeleteOIDCClientCmd())
	commands = append(commands, ListOIDCClientsCmd())

	return commands
}
