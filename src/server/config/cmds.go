package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/shell"
	"google.golang.org/protobuf/encoding/protojson"

	prompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

const (
	listContextHeader = "ACTIVE\tNAME"
)

// returns the active-enterprise-context if set in the config
// otherwise return the active-context if the enterprise license
// is activated on that context's cluster
func deduceActiveEnterpriseContext(ctx context.Context, cfg *config.Config, pachctlCfg *pachctl.Config) (_ string, retErr error) {
	var activeEnterpriseContext string
	if cfg.V2.ActiveEnterpriseContext != "" {
		activeEnterpriseContext = cfg.V2.ActiveEnterpriseContext
	} else {
		c, err := pachctlCfg.NewOnUserMachine(ctx, true)
		if err != nil {
			return "", err
		}
		defer errors.Close(&retErr, c, "close client")
		ctx, cancel := context.WithTimeout(c.Ctx(), time.Second)
		defer cancel()
		state, err := c.Enterprise.GetState(ctx, &enterprise.GetStateRequest{})
		if err != nil {
			return "", errors.EnsureStack(err)
		}
		if state.State == enterprise.State_ACTIVE {
			activeEnterpriseContext = c.ClientContextName()
		}
	}
	return activeEnterpriseContext, nil
}

func ConnectCmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command
	var alias string

	connect := &cobra.Command{
		Use:   "{{alias}} <address>",
		Short: "Connect to a Pachyderm Cluster",
		Long: "This command creates a Pachyderm context at the given address and sets it as active. It stores the pachd address, cluster deployment ID, and actively set project name. \n\n" +
			"If the actively set project no longer exists due to deletion or hard restart / reinstall, you may get an error that can be resolved by setting an existing project (e.g., `default`) to the context. \n" +
			"\t- To list all contexts, use `pachctl config list contexts`. \n" +
			"\t- To view details, use `pachctl config get context <context>`. \n" +
			"\t- To clean up your contexts, use `pachctl config delete context <context>`. \n" +
			"\t- To set a different context as active, use `pachctl config set active-context <context>`. \n" +
			"\t- To set a different project as active, use `pachctl config update context --project foo`.",
		Example: "\t- {{alias}} localhost:80" +
			"\t- {{alias}} localhost:80 --alias my-private-cluster",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			address := args[0]
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}

			contextName := address
			if alias != "" {
				contextName = alias
			}

			context, contextExists := cfg.V2.Contexts[contextName]

			if !contextExists {
				context = new(config.Context)
				pachdAddress, err := grpcutil.ParsePachdAddress(address)
				if err != nil {
					return err
				}
				context.PachdAddress = pachdAddress.Qualified()
				fmt.Printf("New context '%s' created, will connect to Pachyderm at %s\n", contextName, pachdAddress.Qualified())
			}

			cfg.V2.Contexts[contextName] = context
			cfg.V2.ActiveContext = contextName
			fmt.Printf("Context '%s' set as active\n", contextName)

			return cfg.Write()

		}),
	}

	connect.Flags().StringVar(&alias, "alias", "", "Set an alias for the context that is created.")

	commands = append(commands, cmdutil.CreateAlias(connect, "connect"))
	return commands
}

// Cmds returns a slice containing admin commands.
func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	getMetrics := &cobra.Command{
		Short:   "Gets whether metrics are enabled.",
		Long:    "This command returns the status of metric enablement (`pachd.metrics.enabled`).",
		Example: "{{alias}}}",
		Run: cmdutil.Run(func(cmd *cobra.Command, args []string) (retErr error) {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}
			fmt.Printf("%v\n", cfg.V2.Metrics)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(getMetrics, "config get metrics"))

	setMetrics := &cobra.Command{
		Use:     "{{alias}} (true | false)",
		Short:   "Controls whether metrics are enabled or not.",
		Long:    "This command controls whether metrics are enabled or not.",
		Example: "{{alias}} true",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			metrics := true
			if args[0] == "false" {
				metrics = false
			} else if args[0] != "true" {
				return errors.New("invalid argument; use either `true` or `false`")
			}

			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}

			cfg.V2.Metrics = metrics
			return cfg.Write()
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(setMetrics, "config set metrics"))

	getActiveContext := &cobra.Command{
		Short: "Gets the currently active context.",
		Long: "This command returns the currently active context. \n" +
			"\t- To list all contexts, use `pachctl config list contexts`. \n" +
			"\t- To view details, use `pachctl config get context <context>`. \n" +
			"\t- To clean up your contexts, use `pachctl config delete context <context>`. \n" +
			"\t- To set a different context as active, use `pachctl config set active-context <context>`. \n",
		Example: "{{alias}}}",
		Run: cmdutil.Run(func(cmd *cobra.Command, args []string) (retErr error) {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}
			activeContext, _, err := cfg.ActiveContext(false)
			if err != nil {
				return err
			}
			if activeContext == "" {
				fmt.Println("NONE")
			} else {
				fmt.Printf("%s\n", activeContext)
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(getActiveContext, "config get active-context"))

	setActiveContext := &cobra.Command{
		Use:   "{{alias}} <context>",
		Short: "Sets the currently active context.",
		Long: "This command sets the currently active context. This should be a combination of your `proxy.host` value and `proxy.server.http(s)Port number`. \n" +
			"\t- To list all contexts, use `pachctl config list contexts`. \n" +
			"\t- To view details, use `pachctl config get context <context>`. \n" +
			"\t- To clean up your contexts, use `pachctl config delete context <context>`.",
		Example: "{{alias}} grpc://localhost:80",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}
			if _, ok := cfg.V2.Contexts[args[0]]; !ok {
				return errors.Errorf("context does not exist: %s", args[0])
			}
			cfg.V2.ActiveContext = args[0]
			return cfg.Write()
		}),
	}
	shell.RegisterCompletionFunc(setActiveContext, contextCompletion)
	commands = append(commands, cmdutil.CreateAlias(setActiveContext, "config set active-context"))

	getActiveEnterpriseContext := &cobra.Command{
		Short: "Gets the currently active enterprise context.",
		Long:  "This command returns the currently active enterprise context for deployments using Enterprise Server.",
		Run: cmdutil.Run(func(cmd *cobra.Command, args []string) (retErr error) {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}
			activeContext, _, err := cfg.ActiveEnterpriseContext(false)
			if err != nil {
				return err
			}
			if activeContext == "" {
				fmt.Println("NONE")
			} else {
				fmt.Printf("%s\n", activeContext)
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(getActiveEnterpriseContext, "config get active-enterprise-context"))

	setActiveEnterpriseContext := &cobra.Command{
		Use:     "{{alias}} <context>",
		Short:   "Sets the currently active enterprise context.",
		Long:    "This command sets the currently active enterprise context for deployments using Enterprise Server.",
		Example: "{{alias}} foo",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}
			if _, ok := cfg.V2.Contexts[args[0]]; !ok {
				return errors.Errorf("context does not exist: %s", args[0])
			}
			cfg.V2.ActiveEnterpriseContext = args[0]
			return cfg.Write()
		}),
	}
	shell.RegisterCompletionFunc(setActiveEnterpriseContext, contextCompletion)
	commands = append(commands, cmdutil.CreateAlias(setActiveEnterpriseContext, "config set active-enterprise-context"))

	getContext := &cobra.Command{
		Use:     "{{alias}} <context>",
		Short:   "Gets a context.",
		Long:    "This command returns the config of a context by its name. This includes the pachd address, cluster deployment ID, and actively set project name.",
		Example: "{{alias}} foo",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}

			context, ok := cfg.V2.Contexts[args[0]]
			if !ok {
				return errors.Errorf("context does not exist: %s", args[0])
			}

			if err = cmdutil.Encoder("json", os.Stdout).EncodeProto(context); err != nil {
				return errors.EnsureStack(err)
			}

			fmt.Println()
			return nil
		}),
	}
	shell.RegisterCompletionFunc(getContext, contextCompletion)
	commands = append(commands, cmdutil.CreateAlias(getContext, "config get context"))

	var overwrite bool
	setContext := &cobra.Command{
		Use:   "{{alias}} <context>",
		Short: "Set a context.",
		Long:  "This command sets a context config from a given name and a JSON configuration file on stdin",
		Example: "\t- {{alias}} foo" +
			"\t- {{alias}} foo --overwrite",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			name := args[0]

			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}

			if !overwrite {
				if _, ok := cfg.V2.Contexts[name]; ok {
					return errors.Errorf("context '%s' already exists, use `--overwrite` if you wish to replace it", args[0])
				}
			}

			cmdutil.PrintStdinReminder()

			var context config.Context
			in, err := io.ReadAll(os.Stdin)
			if err != nil {
				return errors.Wrap(err, "read stdin")
			}
			if err := protojson.Unmarshal(in, &context); err != nil {
				return errors.Wrapf(err, "malformed context")
			}

			pachdAddress, err := grpcutil.ParsePachdAddress(context.PachdAddress)
			if err != nil {
				if !errors.Is(err, grpcutil.ErrNoPachdAddress) {
					return err
				}
			} else {
				context.PachdAddress = pachdAddress.Qualified()
			}

			cfg.V2.Contexts[name] = &context
			return cfg.Write()
		}),
	}
	setContext.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite a context if it already exists.")
	shell.RegisterCompletionFunc(setContext, contextCompletion)
	commands = append(commands, cmdutil.CreateAlias(setContext, "config set context"))

	var kubeContextName, namespace string
	var enterprise bool
	contextFromKube := &cobra.Command{
		Use:   "{{alias}} <context>",
		Short: "Import a Kubernetes context as a Pachyderm context, and set the active Pachyderm context.",
		Long:  "This command imports a Kubernetes context as a Pachyderm context. By default the current kubernetes context is used.",
		Example: "\t- {{alias}} foo" +
			"\t- {{alias}} foo --overwrite",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			name := args[0]

			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}

			if !overwrite {
				if _, ok := cfg.V2.Contexts[name]; ok {
					return errors.Errorf("context '%s' already exists, use `--overwrite` if you wish to replace it", args[0])
				}
			}

			var context config.Context
			kubeConfig, err := config.RawKubeConfig()
			if err != nil {
				return err
			}

			if kubeContextName == "" {
				kubeContextName = kubeConfig.CurrentContext
			}

			kubeContext := kubeConfig.Contexts[kubeContextName]
			if kubeContext == nil {
				return errors.Errorf("kubernetes context does not exist: %s", kubeContextName)
			}

			if namespace == "" {
				namespace = kubeContext.Namespace
			}

			context = config.Context{
				Source:           config.ContextSource_IMPORTED,
				ClusterName:      kubeContext.Cluster,
				AuthInfo:         kubeContext.AuthInfo,
				Namespace:        namespace,
				EnterpriseServer: enterprise,
			}

			if enterprise {
				cfg.V2.ActiveEnterpriseContext = name
			} else {
				cfg.V2.ActiveContext = name
			}
			cfg.V2.Contexts[name] = &context
			return cfg.Write()
		}),
	}
	contextFromKube.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite a context if it already exists.")
	contextFromKube.Flags().BoolVarP(&enterprise, "enterprise", "e", false, "Configure an enterprise server context.")
	contextFromKube.Flags().StringVarP(&kubeContextName, "kubernetes", "k", "", "Specify the kubernetes context's values to import.")
	contextFromKube.Flags().StringVarP(&namespace, "namespace", "n", "", "Specify a namespace where Pachyderm is deployed.")
	commands = append(commands, cmdutil.CreateAlias(contextFromKube, "config import-kube"))

	var pachdAddress string
	var clusterName string
	var authInfo string
	var serverCAs string
	var project string
	var removeClusterDeploymentID bool
	var updateContext *cobra.Command // standalone declaration so Run() can refer
	updateContext = &cobra.Command{
		Use:   "{{alias}} [<context>]",
		Short: "Updates a context.",
		Long:  "This command updates an existing context config from a given name (or the currently-active context, if no name is given).",
		Example: "\t- {{alias}} foo" +
			"\t- {{alias}} foo --pachd-address localhost:30650" +
			"\t- {{alias}} foo --cluster-name my-cluster" +
			"\t- {{alias}} foo --auth-info my-auth-info" +
			"\t- {{alias}} foo --server-cas /path/to/ca.crt" +
			"\t- {{alias}} foo --namespace my-namespace" +
			"\t- {{alias}} foo --project my-project" +
			"\t- {{alias}} foo --remove-cluster-deployment-id",
		Run: cmdutil.RunBoundedArgs(0, 1, func(cmd *cobra.Command, args []string) (retErr error) {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}

			var context *config.Context
			if len(args) > 0 {
				var ok bool
				context, ok = cfg.V2.Contexts[args[0]]
				if !ok {
					return errors.Errorf("context does not exist: %s", args[0])
				}
			} else {
				var name string
				var err error
				name, context, err = cfg.ActiveContext(true)
				if err != nil {
					return err
				}
				fmt.Printf("editing the currently active context %q\n", name)
			}

			// Use this method since we want to differentiate between no
			// flag being set (the value shouldn't be changed) vs the flag
			// being an empty string (meaning we want to set the value to an
			// empty string)
			if updateContext.Flags().Changed("pachd-address") {
				parsedPachdAddress, err := grpcutil.ParsePachdAddress(pachdAddress)
				if err != nil {
					if errors.Is(err, grpcutil.ErrNoPachdAddress) {
						context.PachdAddress = ""
					} else {
						return err
					}
				} else {
					context.PachdAddress = parsedPachdAddress.Qualified()
				}
			}
			if updateContext.Flags().Changed("cluster-name") {
				context.ClusterName = clusterName
			}
			if updateContext.Flags().Changed("auth-info") {
				context.AuthInfo = authInfo
			}
			if updateContext.Flags().Changed("server-cas") {
				context.ServerCas = serverCAs
			}
			if updateContext.Flags().Changed("namespace") {
				context.Namespace = namespace
			}
			if updateContext.Flags().Changed("project") {
				context.Project = project
			}
			if removeClusterDeploymentID {
				context.ClusterDeploymentId = ""
			}

			return cfg.Write()
		}),
	}
	updateContext.Flags().StringVar(&pachdAddress, "pachd-address", "", "Set a new name pachd address.")
	updateContext.Flags().StringVar(&clusterName, "cluster-name", "", "Set a new cluster name.")
	updateContext.Flags().StringVar(&authInfo, "auth-info", "", "Set a new k8s auth info.")
	updateContext.Flags().StringVar(&serverCAs, "server-cas", "", "Set new trusted CA certs.")
	updateContext.Flags().StringVar(&namespace, "namespace", "", "Set a new namespace.")
	updateContext.Flags().StringVar(&project, "project", "", "Set a new project.")
	updateContext.Flags().BoolVar(&removeClusterDeploymentID, "remove-cluster-deployment-id", false, "Remove the cluster deployment ID field, which will be repopulated on the next 'pachctl' call using this context.")
	shell.RegisterCompletionFunc(updateContext, contextCompletion)
	commands = append(commands, cmdutil.CreateAlias(updateContext, "config update context"))

	deleteContext := &cobra.Command{
		Use:   "{{alias}} <context>",
		Short: "Deletes a context.",
		Long:  "This command deletes a context. It is recommended to clean up your contexts periodically if you have many.",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}
			if _, ok := cfg.V2.Contexts[args[0]]; !ok {
				return errors.Errorf("context does not exist: %s", args[0])
			}
			if cfg.V2.ActiveContext == args[0] {
				return errors.New("cannot delete an active context")
			}
			delete(cfg.V2.Contexts, args[0])
			return cfg.Write()
		}),
	}
	shell.RegisterCompletionFunc(deleteContext, contextCompletion)
	commands = append(commands, cmdutil.CreateAlias(deleteContext, "config delete context"))

	listContext := &cobra.Command{
		Short: "Lists contexts.",
		Long:  "This command lists contexts.",
		Run: cmdutil.Run(func(cmd *cobra.Command, args []string) (retErr error) {
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}

			keys := make([]string, len(cfg.V2.Contexts))
			i := 0
			for key := range cfg.V2.Contexts {
				keys[i] = key
				i++
			}
			sort.Strings(keys)

			activeContext, _, err := cfg.ActiveContext(false)
			if err != nil {
				return err
			}

			activeEnterpriseContext, err := deduceActiveEnterpriseContext(cmd.Context(), cfg, pachctlCfg)
			if err != nil {
				fmt.Printf("Unable to connect with server to deduce enterprise context: %v\n", err.Error())
			}

			fmt.Println(listContextHeader)
			for _, key := range keys {
				var activeMarker string
				if key == activeContext {
					activeMarker = "*"
				}
				var activeEnterpriseMarker string
				if key == activeEnterpriseContext {
					activeEnterpriseMarker = "E"
				}
				fmt.Printf("%s\t%s\n", activeEnterpriseMarker+activeMarker, key)
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(listContext, "config list context"))

	configDocs := &cobra.Command{
		Short: "Manages the pachyderm config.",
		Long:  "Gets/sets pachyderm config values.",
	}
	commands = append(commands, cmdutil.CreateDocsAlias(configDocs, "config", " config "))

	configGetRoot := &cobra.Command{
		Short: "Commands for getting pachyderm config values",
		Long:  "Commands for getting pachyderm config values",
	}
	commands = append(commands, cmdutil.CreateAlias(configGetRoot, "config get"))

	configSetRoot := &cobra.Command{
		Short: "Commands for setting pachyderm config values",
		Long:  "Commands for setting pachyderm config values",
	}
	commands = append(commands, cmdutil.CreateAlias(configSetRoot, "config set"))

	configUpdateRoot := &cobra.Command{
		Short: "Commands for updating pachyderm config values",
		Long:  "Commands for updating pachyderm config values",
	}
	commands = append(commands, cmdutil.CreateAlias(configUpdateRoot, "config update"))

	configDeleteRoot := &cobra.Command{
		Short: "Commands for deleting pachyderm config values",
		Long:  "Commands for deleting pachyderm config values",
	}
	commands = append(commands, cmdutil.CreateAlias(configDeleteRoot, "config delete"))

	configListRoot := &cobra.Command{
		Short: "Commands for listing pachyderm config values",
		Long:  "Commands for listing pachyderm config values",
	}
	commands = append(commands, cmdutil.CreateAlias(configListRoot, "config list"))

	return commands
}

func contextCompletion(_, text string, maxCompletions int64) ([]prompt.Suggest, shell.CacheFunc) {
	cfg, err := config.Read(false, false)
	if err != nil {
		shell.Fatal(err)
	}
	activeContext, _, err := cfg.ActiveContext(false)
	if err != nil {
		shell.Fatal(err)
	}
	var result []prompt.Suggest
	for name, ctx := range cfg.V2.Contexts {
		desc := ctx.PachdAddress
		if name == activeContext {
			desc += " (active)"
		}
		result = append(result, prompt.Suggest{
			Text:        name,
			Description: desc,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		switch {
		case result[i].Text == activeContext:
			return true
		case result[j].Text == activeContext:
			return false
		default:
			return result[i].Text < result[j].Text
		}
	})
	return result, shell.CacheAll
}
