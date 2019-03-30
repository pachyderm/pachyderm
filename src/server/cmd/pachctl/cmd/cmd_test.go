package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"

	"github.com/spf13/cobra"
)

func TestPortForwardError(t *testing.T) {
	os.Setenv("PACHD_ADDRESS", "localhost:30650")
	c := tu.Cmd("pachctl", "version", "--timeout=1ns", "--no-port-forwarding")
	var errMsg bytes.Buffer
	c.Stdout = ioutil.Discard
	c.Stderr = &errMsg
	err := c.Run()
	require.YesError(t, err) // 1ns should prevent even local connections
	require.Matches(t, "port-forward", errMsg.String())
}

func TestNoPort(t *testing.T) {
	os.Setenv("PACHD_ADDRESS", "localhost")
	c := tu.Cmd("pachctl", "version", "--timeout=1ns", "--no-port-forwarding")
	var errMsg bytes.Buffer
	c.Stdout = ioutil.Discard
	c.Stderr = &errMsg
	err := c.Run()
	require.YesError(t, err) // 1ns should prevent even local connections
	require.Matches(t, "30650", errMsg.String())
}

func TestWeirdPortError(t *testing.T) {
	os.Setenv("PACHD_ADDRESS", "localhost:30560")
	c := tu.Cmd("pachctl", "version", "--timeout=1ns", "--no-port-forwarding")
	var errMsg bytes.Buffer
	c.Stdout = ioutil.Discard
	c.Stderr = &errMsg
	err := c.Run()
	require.YesError(t, err) // 1ns should prevent even local connections
	require.Matches(t, "30650", errMsg.String())
}

// Check that no commands have brackets in their names, which indicates that
// 'CreateAlias' was not used properly (or the command just needs to specify
// its name).
func TestCommandAliases(t *testing.T) {
	pachctlCmd := PachctlCmd()

	// Replace the first component with 'pachctl' because it use os.Args[0]
	path := func(cmd *cobra.Command) string {
		return strings.Replace(cmd.CommandPath(), os.Args[0], "pachctl", 1)
	}

	paths := map[string]bool{}

	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		require.True(
			t, cmd.Run != nil || len(cmd.Commands()) > 0,
			"Command is not runnable and has no child commands: '%s' (%s)",
			path(cmd), cmd.Short,
		)

		for _, subcmd := range cmd.Commands() {
			// This should only happen if there is a bug in MergeCommands, or some
			// code is bypassing it.
			require.False(
				t, paths[path(subcmd)],
				"Multiple commands found with the same invocation: '%s'",
				path(subcmd),
			)

			paths[path(subcmd)] = true

			require.True(
				t, subcmd.Short != "",
				"Command must provide a 'Short' description string: '%s'",
				path(subcmd),
			)
			require.True(
				t, subcmd.Long != "",
				"Command must provide a 'Long' description string: '%s' (%s)",
				path(subcmd), subcmd.Short,
			)
			require.False(
				t, strings.ContainsAny(subcmd.Name(), "[<({})>]"),
				"Command name contains invalid characters: '%s' (%s)",
				path(subcmd), subcmd.Short,
			)
			require.True(
				t, subcmd.Use != "",
				"Command must provide a 'Use' string: '%s' (%s)",
				path(subcmd), subcmd.Short,
			)

			walk(subcmd)
		}
	}

	walk(pachctlCmd)
}
