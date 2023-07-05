package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	"github.com/pachyderm/pachyderm/v2/src/server/identityutil"
)

type connectorConfig struct {
	ID      string
	Name    string
	Type    string
	Version int64
	Config  map[string]any
}

func newConnectorConfig(conn *identity.IDPConnector) (*connectorConfig, error) {
	srcConfig, err := identityutil.PickConfig(conn.Config, conn.JsonConfig)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	config := map[string]interface{}{}
	err = json.Unmarshal(srcConfig, &config)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	connConfig := connectorConfig{
		ID:      conn.Id,
		Name:    conn.Name,
		Type:    conn.Type,
		Version: conn.ConfigVersion,
		Config:  config,
	}
	return &connConfig, nil
}

func (c connectorConfig) toIDPConnector() (*identity.IDPConnector, error) {
	cfg, err := structpb.NewStruct(c.Config)
	if err != nil {
		return nil, errors.Wrapf(err, "structpb.NewStruct on %#v", c.Config)
	}
	return &identity.IDPConnector{
		Id:            c.ID,
		Name:          c.Name,
		Type:          c.Type,
		ConfigVersion: c.Version,
		Config:        cfg,
	}, nil
}

func deserializeYAML(file string, target interface{}) error {
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

	return serde.Decode(rawConfigBytes, target)
}

// SetIdentityServerConfigCmd returns a cobra.Command to configure the identity server
func SetIdentityServerConfigCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var file string
	setConfig := &cobra.Command{
		Short: "Set the identity server config",
		Long:  `Set the identity server config`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			var config identity.IdentityServerConfig
			if err := deserializeYAML(file, &config); err != nil {
				return errors.Wrapf(err, "unable to parse config")
			}

			_, err = c.SetIdentityServerConfig(c.Ctx(), &identity.SetIdentityServerConfigRequest{Config: &config})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	setConfig.PersistentFlags().StringVar(&file, "config", "-", `The file to read the YAML-encoded configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(setConfig, "idp set-config")
}

// GetIdentityServerConfigCmd returns a cobra.Command to fetch the current ID server config
func GetIdentityServerConfigCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	getConfig := &cobra.Command{
		Short: "Get the identity server config",
		Long:  `Get the identity server config`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			resp, err := c.GetIdentityServerConfig(c.Ctx(), &identity.GetIdentityServerConfigRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			yamlStr, err := serde.EncodeYAML(resp.Config)
			if err != nil {
				return err
			}
			fmt.Println(string(yamlStr))
			return nil
		}),
	}
	return cmdutil.CreateAlias(getConfig, "idp get-config")
}

// CreateIDPConnectorCmd returns a cobra.Command to create a new IDP integration
func CreateIDPConnectorCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var file string
	createConnector := &cobra.Command{
		Short: "Create a new identity provider connector.",
		Long:  `Create a new identity provider connector.`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			var connector connectorConfig
			if err := deserializeYAML(file, &connector); err != nil {
				return errors.Wrapf(err, "unable to parse config")
			}
			config, err := connector.toIDPConnector()
			if err != nil {
				return err
			}

			_, err = c.CreateIDPConnector(c.Ctx(), &identity.CreateIDPConnectorRequest{Connector: config})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	createConnector.PersistentFlags().StringVar(&file, "config", "-", `The file to read the YAML-encoded connector configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(createConnector, "idp create-connector")
}

// UpdateIDPConnectorCmd returns a cobra.Command to create a new IDP integration
func UpdateIDPConnectorCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var file string
	updateConnector := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Update an existing identity provider connector.",
		Long:  `Update an existing identity provider connector. Only fields which are specified are updated.`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			var connector connectorConfig
			if err := deserializeYAML(file, &connector); err != nil {
				return errors.Wrapf(err, "unable to parse config")
			}

			config, err := connector.toIDPConnector()
			if err != nil {
				return err
			}

			req := &identity.UpdateIDPConnectorRequest{Connector: config}

			_, err = c.UpdateIDPConnector(c.Ctx(), req)
			return grpcutil.ScrubGRPC(err)
		}),
	}
	updateConnector.PersistentFlags().StringVar(&file, "config", "-", `The file to read the YAML-encoded connector configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(updateConnector, "idp update-connector")
}

// GetIDPConnectorCmd returns a cobra.Command to get an IDP connector configuration
func GetIDPConnectorCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	getConnector := &cobra.Command{
		Use:   "{{alias}} <connector id>",
		Short: "Get the config for an identity provider connector.",
		Long:  "Get the config for an identity provider connector.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}

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
			fmt.Println(string(yamlStr))
			return nil
		}),
	}
	return cmdutil.CreateAlias(getConnector, "idp get-connector")
}

// DeleteIDPConnectorCmd returns a cobra.Command to delete an IDP connector
func DeleteIDPConnectorCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	deleteConnector := &cobra.Command{
		Short: "Delete an identity provider connector",
		Long:  "Delete an identity provider connector",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			_, err = c.DeleteIDPConnector(c.Ctx(), &identity.DeleteIDPConnectorRequest{Id: args[0]})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(deleteConnector, "idp delete-connector")
}

// ListIDPConnectorsCmd returns a cobra.Command to list IDP integrations
func ListIDPConnectorsCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	listConnectors := &cobra.Command{
		Short: "List identity provider connectors",
		Long:  `List identity provider connectors`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			resp, err := c.ListIDPConnectors(c.Ctx(), &identity.ListIDPConnectorsRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			for _, conn := range resp.Connectors {
				fmt.Printf("%v - %v (%v)\n", conn.Id, conn.Name, conn.Type)
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(listConnectors, "idp list-connector")
}

// CreateOIDCClientCmd returns a cobra.Command to create a new OIDC client
func CreateOIDCClientCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var file string
	createClient := &cobra.Command{
		Short: "Create a new OIDC client.",
		Long:  `Create a new OIDC client.`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			var client identity.OIDCClient
			if err := deserializeYAML(file, &client); err != nil {
				return errors.Wrapf(err, "unable to parse config")
			}

			resp, err := c.CreateOIDCClient(c.Ctx(), &identity.CreateOIDCClientRequest{Client: &client})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			fmt.Printf("Client secret: %q\n", resp.Client.Secret)
			return nil
		}),
	}
	createClient.PersistentFlags().StringVar(&file, "config", "-", `The file to read the YAML-encoded client configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(createClient, "idp create-client")
}

// DeleteOIDCClientCmd returns a cobra.Command to delete an OIDC client
func DeleteOIDCClientCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	deleteClient := &cobra.Command{
		Use:   "{{alias}} <client ID>",
		Short: "Delete an OIDC client.",
		Long:  `Delete an OIDC client.`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			_, err = c.DeleteOIDCClient(c.Ctx(), &identity.DeleteOIDCClientRequest{Id: args[0]})
			return grpcutil.ScrubGRPC(err)
		}),
	}
	return cmdutil.CreateAlias(deleteClient, "idp delete-client")
}

// GetOIDCClientCmd returns a cobra.Command to get an OIDC client
func GetOIDCClientCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	getClient := &cobra.Command{
		Use:   "{{alias}} <client ID>",
		Short: "Get an OIDC client.",
		Long:  `Get an OIDC client.`,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			resp, err := c.GetOIDCClient(c.Ctx(), &identity.GetOIDCClientRequest{Id: args[0]})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			yamlStr, err := serde.EncodeYAML(resp.Client)
			if err != nil {
				return err
			}
			fmt.Println(string(yamlStr))
			return nil
		}),
	}
	return cmdutil.CreateAlias(getClient, "idp get-client")
}

// UpdateOIDCClientCmd returns a cobra.Command to update an existing OIDC client
func UpdateOIDCClientCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	var file string
	updateClient := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Update an OIDC client.",
		Long:  `Update an OIDC client.`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			var client identity.OIDCClient
			if err := deserializeYAML(file, &client); err != nil {
				return errors.Wrapf(err, "unable to parse config")
			}

			_, err = c.UpdateOIDCClient(c.Ctx(), &identity.UpdateOIDCClientRequest{Client: &client})
			return grpcutil.ScrubGRPC(err)
		}),
	}

	updateClient.PersistentFlags().StringVar(&file, "config", "-", `The file to read the YAML-encoded client configuration from, or '-' for stdin.`)
	return cmdutil.CreateAlias(updateClient, "idp update-client")
}

// ListOIDCClientsCmd returns a cobra.Command to list IDP integrations
func ListOIDCClientsCmd(ctx context.Context, pachctlCfg *pachctl.Config) *cobra.Command {
	listConnectors := &cobra.Command{
		Short: "List OIDC clients.",
		Long:  `List OIDC clients.`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := pachctlCfg.NewOnUserMachine(ctx, true)
			if err != nil {
				return errors.Wrapf(err, "could not connect")
			}
			defer c.Close()

			resp, err := c.ListOIDCClients(c.Ctx(), &identity.ListOIDCClientsRequest{})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			for _, client := range resp.Clients {
				fmt.Printf("%v\n", client.Id)
			}
			return nil
		}),
	}
	return cmdutil.CreateAlias(listConnectors, "idp list-client")
}

// Cmds returns a list of cobra commands for authenticating and authorizing
// users in an auth-enabled Pachyderm cluster.
func Cmds(mainCtx context.Context, pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	idp := &cobra.Command{
		Short: "Commands to manage identity provider integrations",
		Long:  "Commands to manage identity provider integrations",
	}

	commands = append(commands, cmdutil.CreateAlias(idp, "idp"))
	commands = append(commands, GetIdentityServerConfigCmd(mainCtx, pachctlCfg))
	commands = append(commands, SetIdentityServerConfigCmd(mainCtx, pachctlCfg))
	commands = append(commands, CreateIDPConnectorCmd(mainCtx, pachctlCfg))
	commands = append(commands, GetIDPConnectorCmd(mainCtx, pachctlCfg))
	commands = append(commands, UpdateIDPConnectorCmd(mainCtx, pachctlCfg))
	commands = append(commands, DeleteIDPConnectorCmd(mainCtx, pachctlCfg))
	commands = append(commands, ListIDPConnectorsCmd(mainCtx, pachctlCfg))
	commands = append(commands, CreateOIDCClientCmd(mainCtx, pachctlCfg))
	commands = append(commands, GetOIDCClientCmd(mainCtx, pachctlCfg))
	commands = append(commands, UpdateOIDCClientCmd(mainCtx, pachctlCfg))
	commands = append(commands, DeleteOIDCClientCmd(mainCtx, pachctlCfg))
	commands = append(commands, ListOIDCClientsCmd(mainCtx, pachctlCfg))

	return commands
}
