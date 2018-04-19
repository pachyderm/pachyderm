package worker

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"

	prom_api "github.com/prometheus/client_golang/api"
	prom_api_v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom_model "github.com/prometheus/common/model"
)

func TestPrometheusStats(t *testing.T) {

	c := getPachClient(t)
	//	defer require.NoError(t, c.DeleteAll())

	_, err := c.Enterprise.Activate(context.Background(),
		&enterprise.ActivateRequest{ActivationCode: tu.GetTestEnterpriseCode()})
	require.NoError(t, err)
	require.NoError(t, backoff.Retry(func() error {
		resp, err := c.Enterprise.GetState(context.Background(),
			&enterprise.GetStateRequest{})
		if err != nil {
			return err
		}
		if resp.State != enterprise.State_ACTIVE {
			return fmt.Errorf("expected enterprise state to be ACTIVE but was %v", resp.State)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Now that it's activated, run a simple pipeline so we can collect some stats

	dataRepo := tu.UniqueString("TestSimplePipeline_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	commit1, err := c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit1.ID, "file", strings.NewReader("foo"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit1.ID))

	pipeline := tu.UniqueString("TestSimplePipeline")

	_, err = c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pipeline),
			Transform: &pps.Transform{
				Cmd: []string{"bash"},
				Stdin: []string{
					// We want non const runtime vals so histogram rate queries return
					// real results (not NaNs):
					"sleep $(( ( RANDOM % 10 )  + 1 ))",
					// Include case where we will err:
					fmt.Sprintf("touch /pfs/%s/test; if [[ $(cat /pfs/%s/test) == 'fail' ]]; then echo 'Failing'; exit 1; fi", dataRepo, dataRepo),
					fmt.Sprintf("cp /pfs/%s/* /pfs/out/", dataRepo),
				},
			},
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
			Input:        client.NewAtomInput(dataRepo, "/*"),
			OutputBranch: "",
			Update:       false,
			EnableStats:  true,
		},
	)
	require.NoError(t, err)

	_, err = c.FlushCommit([]*pfs.Commit{commit1}, nil)
	require.NoError(t, err)

	// For UI integration, upping this value creates some nice sample data
	numCommits := 5
	var commit *pfs.Commit
	// We already did one successful commit, and we'll add a failing commit
	// below, so we need numCommits-2 more here
	for i := 0; i < numCommits-2; i++ {
		commit, err = c.StartCommit(dataRepo, "master")
		require.NoError(t, err)
		_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("bar"))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	}
	// Now write data that'll make the job fail
	commit, err = c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "test", strings.NewReader("fail"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	_, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)

	// Now hit stats endpoint and see some results
	// get the nodeport of the prometheus deployment
	// kc --namespace=monitoring get svc/prometheus -o json | jq -r .spec.ports[0].nodePort
	port := os.Getenv("PROM_PORT")
	promClient, err := prom_api.NewClient(prom_api.Config{
		Address: fmt.Sprintf("http://127.0.0.1:%v", port),
	})
	require.NoError(t, err)
	promAPI := prom_api_v1.NewAPI(promClient)

	// Datum count queries

	// Prometheus scrapes ~ every 15s, but empirically, we miss data unless I
	// wait this long. This is annoying, and also why this is a single giant
	// test, not many little ones
	time.Sleep(75 * time.Second)
	totalRuns := numCommits + 2 // we have numCommits-1 successfully run datums, but 1 job fails which will run 3 times
	query := fmt.Sprintf("sum(pachyderm_user_datum_count{pipelineName=\"%v\", state=\"started\"})", pipeline)
	result, err := promAPI.Query(context.Background(), query, time.Now())
	require.NoError(t, err)
	resultVec := result.(prom_model.Vector)
	require.Equal(t, 1, len(resultVec))
	require.Equal(t, float64(totalRuns), float64(resultVec[0].Value))

	query = fmt.Sprintf("sum(pachyderm_user_datum_count{pipelineName=\"%v\", state=\"finished\"})", pipeline)
	result, err = promAPI.Query(context.Background(), query, time.Now())
	require.NoError(t, err)
	resultVec = result.(prom_model.Vector)
	require.Equal(t, 1, len(resultVec))
	require.Equal(t, float64(numCommits-1), float64(resultVec[0].Value))

	query = fmt.Sprintf("sum(pachyderm_user_datum_count{pipelineName=\"%v\", state=\"errored\"})", pipeline)
	result, err = promAPI.Query(context.Background(), query, time.Now())
	require.NoError(t, err)
	resultVec = result.(prom_model.Vector)
	require.Equal(t, 1, len(resultVec))
	require.Equal(t, float64(3.0), float64(resultVec[0].Value)) // 3 restarts, one failed job

	// Moving Avg Datum Time Queries
	for _, segment := range []string{"download", "upload"} {
		sum := fmt.Sprintf("rate(pachyderm_user_datum_%v_time_sum{pipelineName=\"%v\"}[5m])", segment, pipeline)
		count := fmt.Sprintf("rate(pachyderm_user_datum_%v_time_count{pipelineName=\"%v\"}[5m])", segment, pipeline)
		query = sum + "/" + count // compute the avg over 5m
		result, err = promAPI.Query(context.Background(), query, time.Now())
		require.NoError(t, err)
		resultVec = result.(prom_model.Vector)
		require.Equal(t, 1, len(resultVec)) // sum gets aggregated no matter how many datums ran
	}
	segment := "proc"
	sum := fmt.Sprintf("rate(pachyderm_user_datum_%v_time_sum{pipelineName=\"%v\"}[5m])", segment, pipeline)
	count := fmt.Sprintf("rate(pachyderm_user_datum_%v_time_count{pipelineName=\"%v\"}[5m])", segment, pipeline)
	query = sum + "/" + count // compute the avg over 5m
	result, err = promAPI.Query(context.Background(), query, time.Now())
	require.NoError(t, err)
	resultVec = result.(prom_model.Vector)
	require.Equal(t, 2, len(resultVec)) // sum gets aggregated no matter how many datums ran, but we get one result for finished datums, and one for errored datums

	// Moving Avg Datum Size Queries
	for _, segment := range []string{"download", "upload"} {
		sum := fmt.Sprintf("rate(pachyderm_user_datum_%v_size_sum{pipelineName=\"%v\"}[5m])", segment, pipeline)
		count := fmt.Sprintf("rate(pachyderm_user_datum_%v_size_count{pipelineName=\"%v\"}[5m])", segment, pipeline)
		query = sum + "/" + count // compute the avg over 5m
		result, err = promAPI.Query(context.Background(), query, time.Now())
		require.NoError(t, err)
		resultVec = result.(prom_model.Vector)
		require.Equal(t, 1, len(resultVec)) // sum gets aggregated no matter how many datums ran
	}

}

// should also write a test to make sure a pipeline that def errs gets counted as such
