package worker

import (
	"fmt"
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
					"sleep $(( ( RANDOM % 10 )  + 1 ))",
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
	numCommits := 1
	var commit *pfs.Commit
	for i := 0; i < numCommits; i++ {
		commit, err = c.StartCommit(dataRepo, "master")
		require.NoError(t, err)
		_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader("bar"))
		require.NoError(t, err)
		require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
	}

	_, err = c.FlushCommit([]*pfs.Commit{commit}, nil)
	require.NoError(t, err)

	// Now hit stats endpoint and see some results
	// get the nodeport of the prometheus deployment
	// kc --namespace=monitoring get svc/prometheus -o json | jq -r .spec.ports[0].nodePort
	port := 31606
	promClient, err := prom_api.NewClient(prom_api.Config{
		Address: fmt.Sprintf("http://127.0.0.1:%v", port),
	})
	require.NoError(t, err)
	promAPI := prom_api_v1.NewAPI(promClient)

	// Datum count queries
	//	result, err := prometheusQuery(prometheusPort, "http://127.0.0.1:31606/api/v1/query?query=sum(pachyderm_user_datum_count{pipelineName=%22TestSimplePipeline0fc7523d1d62%22})")

	time.Sleep(30 * time.Second) // Give prometheus time to scrape the stats

	// a little unclear what the time is for
	query := fmt.Sprintf("pachyderm_user_datum_count{pipelineName=\"%v\"}", pipeline)
	//	countQuery = "pachyderm_user_datum_count"
	result, err := promAPI.Query(context.Background(), query, time.Now())
	require.NoError(t, err)
	fmt.Printf("query result: %v\n", result)
	resultVec := result.(prom_model.Vector)
	require.Equal(t, 2, len(resultVec)) // 2 jobs ran

	query = fmt.Sprintf("sum(pachyderm_user_datum_count{pipelineName=\"%v\", state=\"started\"})", pipeline)
	result, err = promAPI.Query(context.Background(), query, time.Now())
	require.NoError(t, err)
	resultVec = result.(prom_model.Vector)
	require.Equal(t, 1, len(resultVec))
	require.Equal(t, float64(2), float64(resultVec[0].Value))

	query = fmt.Sprintf("sum(pachyderm_user_datum_count{pipelineName=\"%v\", state=\"finished\"})", pipeline)
	result, err = promAPI.Query(context.Background(), query, time.Now())
	require.NoError(t, err)
	resultVec = result.(prom_model.Vector)
	require.Equal(t, 1, len(resultVec))
	require.Equal(t, float64(2), float64(resultVec[0].Value))

	query = fmt.Sprintf("sum(pachyderm_user_datum_count{pipelineName=\"%v\", state=\"errored\"})", pipeline)
	result, err = promAPI.Query(context.Background(), query, time.Now())
	require.NoError(t, err)
	resultVec = result.(prom_model.Vector)
	require.Equal(t, 0, len(resultVec))
	// should check started = 2 finished is 2 and errored is 0

	// Histogram tests
	// REST API
	// http://127.0.0.1:31606/api/v1/query?query=pachyderm_user_datum_proc_time_sum{pipelineName=%22TestSimplePipeline5e95436aba14%22}
	// result: {"status":"success","data":{"resultType":"vector","result":[{"metric":{"app":"pipeline-testsimplepipeline5e95436aba14-v1","component":"worker","instance":"172.17.0.15:9090","job":"kubernetes-endpoints","kubernetes_name":"pipeline-testsimplepipeline5e95436aba14-v1","kubernetes_namespace":"default","pipelineName":"TestSimplePipeline5e95436aba14","state":"finished","suite":"pachyderm","version":"1.7.0-eb187581d79f68e91946590788bcdd925cd70b4f"},"value":[1524094773.735,"NaN"]}]}}

	// data size avg:
	// rate(pachyderm_user_datum_download_size_sum[5m])/rate(pachyderm_user_datum_download_size_count[5m])
	// Note ... im a bit concerned its getting incremented wrong ...

	// datum count
	// Download time
	// Proc time
	// Upload time
	// Download size
	// Upload size

}

// should also write a test to make sure a pipeline that def errs gets counted as such
