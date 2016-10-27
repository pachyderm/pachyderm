package cmd

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	deploycmds "github.com/pachyderm/pachyderm/src/server/pkg/deploy/cmds"
	api "k8s.io/kubernetes/pkg/api/v1"
)

func TestMetrics(t *testing.T) {

	// Run deploy normally, should see METRICS=true
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"deploy", "--dry-run"}
	err := deploycmds.DeployCmd(false).Execute()
	require.NoError(t, err)
	require.NoError(t, w.Close())
	// restore stdout
	os.Stdout = old

	decoder := json.NewDecoder(r)
	foundPachdManifest := false
	for {
		var manifest *api.ReplicationController
		err = decoder.Decode(&manifest)
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		require.NoError(t, err)

		if manifest.ObjectMeta.Name == "pachd" && manifest.Kind == "ReplicationController" {
			foundPachdManifest = true
			falseMetricEnvVar := api.EnvVar{
				Name:  "METRICS",
				Value: "true",
			}
			var env []interface{}
			require.Equal(t, 1, len(manifest.Spec.Template.Spec.Containers))
			for _, value := range manifest.Spec.Template.Spec.Containers[0].Env {
				env = append(env, value)
			}
			require.OneOfEquals(t, interface{}(falseMetricEnvVar), env)
		}
	}
	require.Equal(t, true, foundPachdManifest)

	// Run deploy w dev flag, should see METRICS=false
	r, w, _ = os.Pipe()
	os.Stdout = w

	os.Args = []string{"deploy", "-d", "--dry-run"}
	err = deploycmds.DeployCmd(false).Execute()
	require.NoError(t, err)
	require.NoError(t, w.Close())
	// restore stdout
	os.Stdout = old

	decoder = json.NewDecoder(r)
	foundPachdManifest = false
	for {
		var manifest *api.ReplicationController
		err = decoder.Decode(&manifest)
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if manifest.ObjectMeta.Name == "pachd" && manifest.Kind == "ReplicationController" {
			foundPachdManifest = true
			falseMetricEnvVar := api.EnvVar{
				Name:  "METRICS",
				Value: "false",
			}
			var env []interface{}
			require.Equal(t, 1, len(manifest.Spec.Template.Spec.Containers))
			for _, value := range manifest.Spec.Template.Spec.Containers[0].Env {
				env = append(env, value)
			}
			require.OneOfEquals(t, interface{}(falseMetricEnvVar), env)
		}
	}
	require.Equal(t, true, foundPachdManifest)
}
