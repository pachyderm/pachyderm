// Package cmds implements pachctl commands for the license service.
package cmds

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/spf13/cobra"
)

func ActivateCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	activate := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(activate, "license activate")
}

func AddClusterCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	addCluster := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(addCluster, "license add-cluster")
}

func UpdateClusterCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	updateCluster := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(updateCluster, "license update-cluster")
}

func DeleteClusterCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	deleteCluster := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(deleteCluster, "license delete-cluster")
}

func ListClustersCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	listClusters := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(listClusters, "license list-clusters")
}

func DeleteAllCmd(pachctlCfg *pachctl.Config) *cobra.Command {
	deleteAll := &cobra.Command{
		Use:   "{{alias}}",
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			return errors.New("Deprecated. Please don't use.")
		}),
	}
	return cmdutil.CreateAlias(deleteAll, "license delete-all")
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
	return cmdutil.CreateAlias(getState, "license get-state")
}

func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	enterprise := &cobra.Command{
		Short: "Deprecated. Please don't use.",
		Long:  "Deprecated. Please don't use.",
	}
	commands = append(commands, cmdutil.CreateAlias(enterprise, "license"))
	commands = append(commands, ActivateCmd(pachctlCfg))
	commands = append(commands, AddClusterCmd(pachctlCfg))
	commands = append(commands, UpdateClusterCmd(pachctlCfg))
	commands = append(commands, DeleteClusterCmd(pachctlCfg))
	commands = append(commands, ListClustersCmd(pachctlCfg))
	commands = append(commands, DeleteAllCmd(pachctlCfg))
	commands = append(commands, GetStateCmd(pachctlCfg))

	return commands
}
