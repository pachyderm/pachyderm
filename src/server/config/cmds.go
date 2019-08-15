package cmds

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/spf13/cobra"
)

const (
	listContextHeader = "ACTIVE\tNAME"
)

func readContext() (*config.Context, error) {
	var buf bytes.Buffer
	var decoder *json.Decoder
	var result config.Context

	contextReader := io.TeeReader(os.Stdin, &buf)
	fmt.Println("Reading from stdin.")
	decoder = json.NewDecoder(contextReader)

	if err := jsonpb.UnmarshalNext(decoder, &result); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("malformed context: %s", err)
	}
	return &result, nil
}

// Cmds returns a slice containing admin commands.
func Cmds() []*cobra.Command {
	marshaller := &jsonpb.Marshaler{
		Indent:   "  ",
		OrigName: true,
	}

	var commands []*cobra.Command

	getMetrics := &cobra.Command{
		Short: "Gets whether metrics are enabled.",
		Long:  "Gets whether metrics are enabled.",
		Run: cmdutil.Run(func(args []string) (retErr error) {
			cfg, err := config.Read()
			if err != nil {
				return err
			}
			fmt.Printf("%v\n", cfg.V2.Metrics)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(getMetrics, "config get metrics"))

	setMetrics := &cobra.Command{
		Short: "Sets whether metrics are enabled.",
		Long:  "Sets whether metrics are enabled.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			metrics := true
			if args[0] == "false" {
				metrics = false
			} else if args[0] != "true" {
				return errors.New("invalid argument; use either `true` or `false`")
			}

			cfg, err := config.Read()
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
		Long:  "Gets the currently active context.",
		Run: cmdutil.Run(func(args []string) (retErr error) {
			cfg, err := config.Read()
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", cfg.V2.ActiveContext)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(getActiveContext, "config get active-context"))

	setActiveContext := &cobra.Command{
		Short: "Sets the currently active context.",
		Long:  "Sets the currently active context.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			cfg, err := config.Read()
			if err != nil {
				return err
			}
			if _, ok := cfg.V2.Contexts[args[0]]; !ok {
				return fmt.Errorf("context does not exist: %s", args[0])
			}
			cfg.V2.ActiveContext = args[0]
			return cfg.Write()
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(setActiveContext, "config set active-context"))

	getContext := &cobra.Command{
		Short: "Gets a context.",
		Long:  "Gets the config of a context by its name.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			cfg, err := config.Read()
			if err != nil {
				return err
			}

			context, ok := cfg.V2.Contexts[args[0]]
			if !ok {
				return fmt.Errorf("context does not exist: %s", args[0])
			}

			if err = marshaller.Marshal(os.Stdout, context); err != nil {
				return err
			}

			fmt.Println()
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(getContext, "config get context"))

	var overwrite bool
	setContext := &cobra.Command{
		Short: "Set a context.",
		Long:  "Set a context config from a given name and JSON stdin.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			cfg, err := config.Read()
			if err != nil {
				return err
			}

			context, err := readContext()
			if err != nil {
				return err
			}

			if !overwrite {
				if _, ok := cfg.V2.Contexts[args[0]]; ok {
					return fmt.Errorf("context '%s' already exists, use `--overwrite` if you wish to replace it", args[0])
				}
			}

			cfg.V2.Contexts[args[0]] = context
			return cfg.Write()
		}),
	}
	setContext.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite a context if it already exists.")
	commands = append(commands, cmdutil.CreateAlias(setContext, "config set context"))

	var pachdAddress string
	updateContext := &cobra.Command{
		Short: "Updates a context.",
		Long:  "Updates an existing context config from a given name.",
		Run: cmdutil.RunCmdFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			cfg, err := config.Read()
			if err != nil {
				return err
			}

			context, ok := cfg.V2.Contexts[args[0]]
			if !ok {
				return fmt.Errorf("context does not exist: %s", args[0])
			}

			if cmd.Flags().Changed("pachd-address") {
				// Use this method since we want to differentiate between no
				// `pachd-address` flag being set (the value shouldn't be
				// changed) vs the flag being an empty string (meaning we want
				// to set the value to an empty string)
				context.PachdAddress = pachdAddress
			}

			return cfg.Write()
		}),
	}
	updateContext.Flags().StringVar(&pachdAddress, "pachd-address", "", "Set a new name pachd address.")
	commands = append(commands, cmdutil.CreateAlias(updateContext, "config update context"))

	deleteContext := &cobra.Command{
		Short: "Deletes a context.",
		Long:  "Deletes a context.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			cfg, err := config.Read()
			if err != nil {
				return err
			}
			if _, ok := cfg.V2.Contexts[args[0]]; !ok {
				return fmt.Errorf("context does not exist: %s", args[0])
			}
			if cfg.V2.ActiveContext == args[0] {
				return errors.New("cannot delete an active context")
			}
			delete(cfg.V2.Contexts, args[0])
			return cfg.Write()
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(deleteContext, "config delete context"))

	listContext := &cobra.Command{
		Short: "Lists contexts.",
		Long:  "Lists contexts.",
		Run: cmdutil.Run(func(args []string) (retErr error) {
			cfg, err := config.Read()
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

			fmt.Println(listContextHeader)
			for _, key := range keys {
				if key == cfg.V2.ActiveContext {
					fmt.Printf("*\t%s\n", key)
				} else {
					fmt.Printf("\t%s\n", key)
				}
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(listContext, "config list context"))

	configDocs := &cobra.Command{
		Short: "Manages the pachyderm config.",
		Long:  "Gets/sets pachyderm config values.",
	}
	commands = append(commands, cmdutil.CreateDocsAlias(configDocs, "config", "^pachctl config "))

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
