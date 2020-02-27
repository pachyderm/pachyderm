package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/serde"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func YAMLToJSONString(t *testing.T, yamlStr string) string {
	holder := make(map[string]interface{})
	if err := serde.DecodeYAML([]byte(yamlStr), &holder); err != nil {
		t.Fatalf("error parsing TFJob: %v", err)
	}
	result, err := json.Marshal(holder)
	if err != nil {
		t.Fatalf("error marshalling TFJob to JSON: %v", err)
	}
	return string(result)
}

func TestTFJobBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c := tu.GetPachClient(t)
	require.NoError(t, c.DeleteAll())

	dataRepo := tu.UniqueString("TestSimplePipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	pipeline := tu.UniqueString("pipeline1")
	tfJobString := YAMLToJSONString(t, `
    apiVersion: kubeflow.org/v1
    kind: TFJob
    metadata:
      generateName: tfjob
      namespace: kubeflow
    spec:
      tfReplicaSpecs:
        PS:
          replicas: 1
          restartPolicy: OnFailure
          template:
            spec:
              containers:
              - name: tensorflow
                image: gcr.io/your-project/your-image
                command:
                  - python
                  - -m
                  - trainer.task
                  - --batch_size=32
                  - --training_steps=1000
        Worker:
          replicas: 3
          restartPolicy: OnFailure
          template:
            spec:
              containers:
              - name: tensorflow
                image: gcr.io/your-project/your-image
                command:
                  - python
                  - -m
                  - trainer.task
                  - --batch_size=32
                  - --training_steps=1000
    `)
	_, err := c.PpsAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Input:    client.NewPFSInput(dataRepo, "/*"),
			TFJob:    &pps.TFJob{TFJob: tfJobString},
		})
	require.YesError(t, err)
	require.Matches(t, "not supported yet", err.Error())
}
