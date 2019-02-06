package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

var stdoutMutex = &sync.Mutex{}

func TestMetricsNormalDeployment(t *testing.T) {
	// Run deploy normally, should see METRICS=true
	testDeploy(t, false, true, true)
}

func TestMetricsNormalDeploymentNoMetricsFlagSet(t *testing.T) {
	// Run deploy normally, should see METRICS=true
	testDeploy(t, false, false, false)
}

func TestMetricsDevDeployment(t *testing.T) {
	// Run deploy w dev flag, should see METRICS=false
	testDeploy(t, true, true, false)
}

func TestMetricsDevDeploymentNoMetricsFlagSet(t *testing.T) {
	// Run deploy w dev flag, should see METRICS=false
	testDeploy(t, true, false, false)
}

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

func TestNoPortError(t *testing.T) {
	os.Setenv("PACHD_ADDRESS", "127.127.127.0")
	c := tu.Cmd("pachctl", "version", "--timeout=1ns", "--no-port-forwarding")
	var errMsg bytes.Buffer
	c.Stdout = ioutil.Discard
	c.Stderr = &errMsg
	err := c.Run()
	require.YesError(t, err) // 1ns should prevent even local connections
	require.Matches(t, "does not seem to be host:port", errMsg.String())
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
