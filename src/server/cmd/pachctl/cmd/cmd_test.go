package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	api "k8s.io/kubernetes/pkg/api/v1"
)

var stdoutMutex = &sync.Mutex{}

func TestMetricsNormalDeployment(t *testing.T) {
	// Run deploy normally, should see METRICS=true
	testDeploy(t, false, false, true)
}

func TestMetricsNormalDeploymentNoMetricsFlagSet(t *testing.T) {
	// Run deploy normally, should see METRICS=true
	testDeploy(t, false, true, false)
}

func TestMetricsDevDeployment(t *testing.T) {
	// Run deploy w dev flag, should see METRICS=false
	testDeploy(t, true, false, false)
}

func TestMetricsDevDeploymentNoMetricsFlagSet(t *testing.T) {
	// Run deploy w dev flag, should see METRICS=false
	testDeploy(t, true, true, false)
}

func testDeploy(t *testing.T, devFlag bool, noMetrics bool, expectedEnvValue bool) {
	t.Parallel()
	stdoutMutex.Lock()
	defer stdoutMutex.Unlock()

	// Setup user config prior to test
	// So that stdout only contains JSON no warnings
	_, err := config.Read()
	require.NoError(t, err)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{
		"pachctl",
		"deploy",
		"local",
		"--dry-run",
		fmt.Sprintf("-d=%v", devFlag),
		fmt.Sprintf("--no-metrics=%v", noMetrics),
	}
	rootCommand, err := PachctlCmd("127.0.0.1:30650")
	require.NoError(t, err)
	err = rootCommand.Execute()
	require.NoError(t, err)
	require.NoError(t, w.Close())
	// restore stdout
	os.Stdout = old

	decoder := json.NewDecoder(r)
	foundPachdManifest := false
	// Loop through generated manifest until we find a
	// ReplicationController (limit of 100 makes sure test
	// fails quickly if there is no RC)
	for i := 0; i < 100; i++ {
		var manifest *api.ReplicationController
		err = decoder.Decode(&manifest)
		if err == io.EOF {
			break
		}
		if err != nil {
			// Not a replication controller
			continue
		}
		require.NoError(t, err)
		if manifest.ObjectMeta.Name == "pachd" && manifest.Kind == "ReplicationController" {
			foundPachdManifest = true
			expectedMetricEnvVar := api.EnvVar{
				Name:  "METRICS",
				Value: fmt.Sprintf("%v", expectedEnvValue),
			}
			var env []interface{}
			require.Equal(t, 1, len(manifest.Spec.Template.Spec.Containers))
			for _, value := range manifest.Spec.Template.Spec.Containers[0].Env {
				env = append(env, value)
			}
			require.OneOfEquals(t, interface{}(expectedMetricEnvVar), env)
		}
	}
	require.Equal(t, true, foundPachdManifest)
}
