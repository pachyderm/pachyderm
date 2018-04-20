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

	pipeline := tu.UniqueString("TestSimplePipeline")
	// We want several commits (for multiple jobs) and several datums per job
	// For semi meaningful time series results
	numCommits := 5
	numDatums := 10

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
				Constant: uint64(numDatums),
			},
			Input:        client.NewAtomInput(dataRepo, "/*"),
			OutputBranch: "",
			Update:       false,
			EnableStats:  true,
		},
	)
	require.NoError(t, err)

	var commit *pfs.Commit
	// Do numCommits-1 commits w good data
	for i := 0; i < numCommits-1; i++ {
		commit, err = c.StartCommit(dataRepo, "master")
		require.NoError(t, err)
		// We want several datums per job so that we have multiple data points
		// per job time series
		for j := 0; j < numDatums; j++ {
			_, err = c.PutFile(dataRepo, commit.ID, fmt.Sprintf("file%v", j), strings.NewReader("bar"))
			require.NoError(t, err)
		}
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
		// Prometheus scrapes every 10s
		// We run a new job outside this window so that we see a more organic
		// time series
		time.Sleep(15 * time.Second)
	}
	// Now write data that'll make the job fail
	commit, err = c.StartCommit(dataRepo, "master")
	require.NoError(t, err)
	_, err = c.PutFile(dataRepo, commit.ID, "test", strings.NewReader("fail"))
	require.NoError(t, err)
	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))

	_, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)

	port := os.Getenv("PROM_PORT")
	promClient, err := prom_api.NewClient(prom_api.Config{
		Address: fmt.Sprintf("http://127.0.0.1:%v", port),
	})
	require.NoError(t, err)
	promAPI := prom_api_v1.NewAPI(promClient)

	// Prometheus scrapes ~ every 15s, but empirically, we miss data unless I
	// wait this long. This is annoying, and also why this is a single giant
	// test, not many little ones
	time.Sleep(45 * time.Second)

	datumCountQuery := func(t *testing.T, query string, expected float64) {
		result, err := promAPI.Query(context.Background(), query, time.Now())
		require.NoError(t, err)
		resultVec := result.(prom_model.Vector)
		require.Equal(t, 1, len(resultVec))
		require.Equal(t, expected, float64(resultVec[0].Value))
	}
	// Datum count queries
	t.Run("DatumCountStarted", func(t *testing.T) {
		query := fmt.Sprintf("sum(pachyderm_user_datum_count{pipelineName=\"%v\", state=\"started\"})", pipeline)
		datumCountQuery(t, query, float64((numCommits-1)*numDatums+3)) // 3 extra for failed datum restarts on the last job
	})

	t.Run("DatumCountFinished", func(t *testing.T) {
		query := fmt.Sprintf("sum(pachyderm_user_datum_count{pipelineName=\"%v\", state=\"finished\"})", pipeline)
		datumCountQuery(t, query, float64((numCommits-1)*numDatums))
	})

	t.Run("DatumCountErrored", func(t *testing.T) {
		query := fmt.Sprintf("sum(pachyderm_user_datum_count{pipelineName=\"%v\", state=\"errored\"})", pipeline)
		datumCountQuery(t, query, float64(3.0))
	})

	// Test queries across all jobs
	filter := "(instance,exported_job)"
	// instance is an auto recorded label w the IP of the pod ... this will
	// become helpful when debugging certain workers

	// Avg Datum Time Queries
	avgDatumQuery := func(t *testing.T, sumQuery string, countQuery string, expected int) {
		query := "(" + sumQuery + ")/(" + countQuery + ")"
		result, err := promAPI.Query(context.Background(), query, time.Now())
		require.NoError(t, err)
		resultVec := result.(prom_model.Vector)
		require.Equal(t, expected, len(resultVec))
	}
	for _, segment := range []string{"download", "upload"} {
		t.Run(fmt.Sprintf("AcrossJobsDatumTime=%v", segment), func(t *testing.T) {
			sum := fmt.Sprintf("sum(pachyderm_user_datum_%v_time_sum{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
			count := fmt.Sprintf("sum(pachyderm_user_datum_%v_time_count{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
			avgDatumQuery(t, sum, count, 1)
		})
	}
	segment := "proc"
	t.Run(fmt.Sprintf("AcrossJobsDatumTime=%v", segment), func(t *testing.T) {
		sum := fmt.Sprintf("sum(pachyderm_user_datum_%v_time_sum{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
		count := fmt.Sprintf("sum(pachyderm_user_datum_%v_time_count{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
		// sum gets aggregated no matter how many datums ran, but we get one result for finished datums, and one for errored datums
		avgDatumQuery(t, sum, count, 2)
	})

	// Avg Datum Size Queries
	for _, segment := range []string{"download", "upload"} {
		t.Run(fmt.Sprintf("AcrossJobsDatumSize=%v", segment), func(t *testing.T) {
			sum := fmt.Sprintf("sum(pachyderm_user_datum_%v_size_sum{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
			count := fmt.Sprintf("sum(pachyderm_user_datum_%v_size_count{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
			avgDatumQuery(t, sum, count, 1)
		})
	}

	// Now test aggregating per job
	filter = "(instance)"

	// Avg Datum Time Queries
	expectedCounts := map[string]int{
		"download": numCommits + 1, // we expect 5 jobs, plus there's always an extra value w no job label
		"upload":   numCommits - 1, //Since 1 job failed, there will only be 4 upload times
	}
	for _, segment := range []string{"download", "upload"} {
		t.Run(fmt.Sprintf("PerJobDatumTime=%v", segment), func(t *testing.T) {
			sum := fmt.Sprintf("sum(pachyderm_user_datum_%v_time_sum{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
			count := fmt.Sprintf("sum(pachyderm_user_datum_%v_time_count{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
			avgDatumQuery(t, sum, count, expectedCounts[segment])
		})
	}
	segment = "proc"
	t.Run(fmt.Sprintf("PerJobDatumTime=%v", segment), func(t *testing.T) {
		sum := fmt.Sprintf("sum(pachyderm_user_datum_%v_time_sum{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
		count := fmt.Sprintf("sum(pachyderm_user_datum_%v_time_count{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
		avgDatumQuery(t, sum, count, numCommits)
	})

	// Avg Datum Size Queries
	expectedCounts = map[string]int{
		"download": numCommits - 1, // Download size gets reported after job completion, and one job fails
		"upload":   numCommits - 1, //Since 1 job failed, there will only be 4 upload times
	}
	for _, segment := range []string{"download", "upload"} {
		t.Run(fmt.Sprintf("PerJobDatumSize=%v", segment), func(t *testing.T) {
			sum := fmt.Sprintf("sum(pachyderm_user_datum_%v_size_sum{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
			count := fmt.Sprintf("sum(pachyderm_user_datum_%v_size_count{pipelineName=\"%v\"}) without %v", segment, pipeline, filter)
			avgDatumQuery(t, sum, count, expectedCounts[segment])
		})
	}
}
