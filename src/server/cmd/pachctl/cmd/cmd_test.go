package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	deploycmds "github.com/pachyderm/pachyderm/src/server/pkg/deploy/cmds"
	helpers "github.com/pachyderm/pachyderm/src/server/pkg/testing"
	"github.com/spf13/cobra"
	api "k8s.io/kubernetes/pkg/api/v1"
)

func rootCmd(t *testing.T) *cobra.Command {
	cmd, err := PachctlCmd("0.0.0.0:30650")
	require.NoError(t, err)
	return cmd
}

func TestPachctl(t *testing.T) {
	c, err := client.NewFromAddress("0.0.0.0:30650")
	require.NoError(t, err)
	cmd := rootCmd(t)
	expectedState := &helpers.State{}
	repoName := helpers.UniqueString("TestPachctl")

	repoState := expectedState.Repo(repoName)
	t.Run("create-repo", func(t *testing.T) {
		helpers.TestCmd(cmd, []string{"create-repo", repoName}, "", expectedState, c, t)
	})

	commitState1 := repoState.Commit("master/0")
	t.Run("start/finish-commit", func(t *testing.T) {
		commitState1.Info.CommitType = client.CommitTypeWrite
		helpers.TestCmd(cmd, []string{"start-commit", repoName, "master"}, "", expectedState, c, t)
		fileState1 := commitState1.File("/file1")
		fileState1.Content = "foo\n"
		helpers.TestCmd(cmd, []string{"put-file", repoName, "master", "file1"}, fileState1.Content, expectedState, c, t)
		commitState1.Info.CommitType = client.CommitTypeRead
		helpers.TestCmd(cmd, []string{"finish-commit", repoName, "master"}, "", expectedState, c, t)
	})

	commitState2 := repoState.Commit("master/1")
	fileState2 := commitState2.File("/file2")
	fileState2.Content = "bar\n"
	t.Run("put-file", func(t *testing.T) {
		helpers.TestCmd(cmd, []string{"put-file", repoName, "master", "file2", "-c"}, fileState2.Content, expectedState, c, t)
	})

	pipelineName := helpers.UniqueString("TestPachctlPipeline")
	_, outputRepoState := expectedState.Pipeline(pipelineName)
	outCommitState1 := outputRepoState.Commit("")
	outCommitState1.Info.Provenance = append(outCommitState1.Info.Provenance, commitState1.Info.Commit)
	outCommitState1.File("/file1").Content = "foo\n"
	outCommitState2 := outputRepoState.Commit("")
	outCommitState2.Info.Provenance = append(outCommitState2.Info.Provenance, commitState2.Info.Commit)
	outCommitState2.File("/file1").Content = "foo\n"
	t.Run("create-pipeline", func(t *testing.T) {
		createPipelineRequest := &pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipelineName),
			Transform: &pps.Transform{
				Cmd: []string{"cp", fmt.Sprintf("/pfs/%s/file1", repoName), "/pfs/out/file1"},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Strategy: pps.ParallelismSpec_CONSTANT,
				Constant: 1,
			},
			Inputs: []*pps.PipelineInput{
				{
					Repo: client.NewRepo(repoName),
					Method: &pps.Method{
						Partition:   pps.Partition_BLOCK,
						Incremental: pps.Incremental_NONE,
					},
				},
			},
		}
		marshaler := &jsonpb.Marshaler{}
		pipelineSpec, err := marshaler.MarshalToString(createPipelineRequest)
		require.NoError(t, err)
		helpers.TestCmd(cmd, []string{"create-pipeline"}, pipelineSpec, expectedState, c, t)
	})
}

func TestMetrics(t *testing.T) {

	// Run deploy normally, should see METRICS=true
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = []string{"deploy", "--dry-run"}
	err := deploycmds.DeployCmd().Execute()
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
	err = deploycmds.DeployCmd().Execute()
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
