package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
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
