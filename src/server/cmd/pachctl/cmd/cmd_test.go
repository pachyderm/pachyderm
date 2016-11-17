package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	deploycmds "github.com/pachyderm/pachyderm/src/server/pkg/deploy/cmds"
	api "k8s.io/kubernetes/pkg/api/v1"
)

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
	fmt.Printf("running testDeploy: dev %v, nometrics %v, expectedval %v\n", devFlag, noMetrics, expectedEnvValue)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{
		"deploy",
		"local",
		"--dry-run",
		fmt.Sprintf("-d=%v", devFlag),
	}
	// the noMetrics flag is defined globally, so is undefined on just this command
	// but we can pass it in directly to the command:
	err := deploycmds.DeployCmd(!noMetrics).Execute()
	require.NoError(t, err)
	require.NoError(t, w.Close())
	// restore stdout
	os.Stdout = old

	b := make([]byte, 100000)
	n, err := r.Read(b)
	b = b[0:n]
	jsonReader := bytes.NewBuffer(b)
	fmt.Printf("got result [%v]\n", string(b))

	decoder := json.NewDecoder(jsonReader)
	foundPachdManifest := false
	// Loop through generated manifest until we find a
	// ReplicationController (limit of 100 makes sure test
	// fails quickly if there is no RC)
	for i := 0; i < 100; i++ {
		var manifest *api.ReplicationController
		err = decoder.Decode(&manifest)
		if err == io.EOF {
			fmt.Printf("got EOF\n")
			break
		}
		if err != nil {
			// Not a replication controller
			fmt.Printf("Got error decoding: %v\n", err)
			continue
		}
		require.NoError(t, err)
		fmt.Printf("this manifest obj: %v\n", manifest)
		fmt.Printf("this name: %v\n", manifest.ObjectMeta.Name)
		fmt.Printf("this kind: %v\n\n", manifest.Kind)
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
