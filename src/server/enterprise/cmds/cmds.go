// Package cmds implements pachctl commands for the enterprise service.
package cmds

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/spf13/cobra"
)

func DeactivateCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	deactivate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}

	return cmdutil.CreateAlias(deactivate, "enterprise deactivate")
}

func RegisterCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	register := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}

	return cmdutil.CreateAlias(register, "enterprise register")
}

func GetStateCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	getState := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}

	return cmdutil.CreateAlias(getState, "enterprise get-state")
}

func SyncContextsCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	syncContexts := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(syncContexts, "enterprise sync-contexts")
}

func HeartbeatCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	heartbeat := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(heartbeat, "enterprise heartbeat")
}

func PauseCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	pause := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(pause, "enterprise pause")
}

func UnpauseCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	unpause := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(unpause, "enterprise unpause")
}

func PauseStatusCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	pauseStatus := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(pauseStatus, "enterprise pause-status")
}

// Cmds returns pachctl commands related to Pachyderm Enterprise
func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	enterprise := &cobra.Command{
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
	}
	commands = append(commands, cmdutil.CreateAlias(enterprise, "enterprise"))

	commands = append(commands, RegisterCmd(pachctlCfg))
	commands = append(commands, DeactivateCmd(pachctlCfg))
	commands = append(commands, GetStateCmd(pachctlCfg))
	commands = append(commands, SyncContextsCmd(pachctlCfg))
	commands = append(commands, HeartbeatCmd(pachctlCfg))
	commands = append(commands, PauseCmd(pachctlCfg))
	commands = append(commands, UnpauseCmd(pachctlCfg))
	commands = append(commands, PauseStatusCmd(pachctlCfg))

	return commands
}
