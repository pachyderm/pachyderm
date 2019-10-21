package driver

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"reflect"
	"sync"
	"testing"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/prometheus/client_golang/prometheus"
	prometheus_proto "github.com/prometheus/client_model/go"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

var etcdClient *etcd.Client
var etcdOnce sync.Once

func getEtcdClient(t *testing.T) *etcd.Client {
	// src/server/pfs/server/driver.go expects an etcd server at "localhost:32379"
	// Try to establish a connection before proceeding with the test (which will
	// fail if the connection can't be established)
	etcdAddress := "localhost:32379"
	etcdOnce.Do(func() {
		require.NoError(t, backoff.Retry(func() error {
			var err error
			etcdClient, err = etcd.New(etcd.Config{
				Endpoints:   []string{etcdAddress},
				DialOptions: client.DefaultDialOptions(),
			})
			if err != nil {
				return fmt.Errorf("could not connect to etcd: %s", err.Error())
			}
			return nil
		}, backoff.NewTestingBackOff()))
	})
	return etcdClient
}

var pachClient *client.APIClient
var getPachClientOnce sync.Once

func getPachClient(t testing.TB) *client.APIClient {
	getPachClientOnce.Do(func() {
		var err error
		if addr := os.Getenv("PACHD_PORT_650_TCP_ADDR"); addr != "" {
			pachClient, err = client.NewInCluster()
		} else {
			pachClient, err = client.NewForTest()
		}
		require.NoError(t, err)
	})
	return pachClient
}

var inputRepo = "inputRepo"
var testPipelineInfo = &pps.PipelineInfo{
	Pipeline: client.NewPipeline("testPipeline"),
	Transform: &pps.Transform{
		Cmd: []string{"cp", path.Join("/pfs", inputRepo, "file"), "/pfs/out/file"},
	},
	ParallelismSpec: &pps.ParallelismSpec{
		Constant: 1,
	},
	ResourceRequests: &pps.ResourceSpec{
		Memory: "100M",
		Cpu:    0.5,
	},
	Input: client.NewPFSInput(inputRepo, "/*"),
}

func newTestDriver(t *testing.T) *driver {
	d, err := NewDriver(testPipelineInfo, getPachClient(t), tu.GetKubeClient(t), getEtcdClient(t), tu.UniqueString("driverTest"))
	require.NoError(t, err)
	return d.(*driver)
}

func TestGetExpectedNumWorkers(t *testing.T) {
	d := newTestDriver(t)

	// An empty parallelism spec should default to 1 worker
	d.pipelineInfo.ParallelismSpec.Constant = 0
	d.pipelineInfo.ParallelismSpec.Coefficient = 0
	workers, err := d.GetExpectedNumWorkers()
	require.NoError(t, err)
	require.Equal(t, 1, workers)

	// A constant should literally be returned
	d.pipelineInfo.ParallelismSpec.Constant = 1
	workers, err = d.GetExpectedNumWorkers()
	require.NoError(t, err)
	require.Equal(t, 1, workers)

	d.pipelineInfo.ParallelismSpec.Constant = 3
	workers, err = d.GetExpectedNumWorkers()
	require.NoError(t, err)
	require.Equal(t, 3, workers)

	// Constant and Coefficient cannot both be non-zero
	d.pipelineInfo.ParallelismSpec.Coefficient = 0.5
	workers, err = d.GetExpectedNumWorkers()
	require.YesError(t, err)

	// No parallelism spec should default to 1 worker
	d.pipelineInfo.ParallelismSpec = nil
	workers, err = d.GetExpectedNumWorkers()
	require.NoError(t, err)
	require.Equal(t, 1, workers)

	// TODO: test a non-zero coefficient - requires setting up a number of nodes with the kubeClient
}

func requireLogs(t *testing.T, pattern string, cb func(logs.TaggedLogger)) {
	logger := logs.NewMockLogger()
	buffer := &bytes.Buffer{}
	logger.Writer = buffer
	logger.Job = "job-id"

	cb(logger)

	result := string(buffer.Bytes())

	if pattern == "" {
		require.Equal(t, "", result, "callback should not have logged anything")
	} else {
		require.Matches(t, pattern, result, "callback did not log the expected message")
	}
}

func requireMetric(t *testing.T, metric prometheus.Collector, labels []string, cb func(prometheus_proto.Metric)) {
	reg := prometheus.NewRegistry()
	require.NoError(t, reg.Register(metric))

	stats, err := reg.Gather()
	require.NoError(t, err)

	// Add a placeholder for the state label even if it isn't used
	for len(labels) < 3 {
		labels = append(labels, "")
	}

	// We only have one metric in the registry, so skip over the family level
	for _, family := range stats {
		for _, metric := range family.Metric {
			var pipeline, job, state string
			for _, pair := range metric.Label {
				switch *pair.Name {
				case "pipeline":
					pipeline = *pair.Value
				case "job":
					job = *pair.Value
				case "state":
					state = *pair.Value
				default:
					require.True(t, false, fmt.Sprintf("unexpected metric label: %s", *pair.Name))
				}
			}

			metricLabels := []string{pipeline, job, state}
			if reflect.DeepEqual(labels, metricLabels) {
				cb(*metric)
				return
			}
		}
	}

	require.True(t, false, fmt.Sprintf("no matching metric found for labels: %v", labels))
}

func requireCounter(t *testing.T, counter *prometheus.CounterVec, labels []string, value float64) {
	requireMetric(t, counter, labels, func(m prometheus_proto.Metric) {
		require.NotNil(t, m.Counter)
		require.Equal(t, value, *m.Counter.Value)
	})
}

func requireHistogram(t *testing.T, histogram *prometheus.HistogramVec, labels []string, value float64) {
	requireMetric(t, histogram, labels, func(m prometheus_proto.Metric) {
		require.NotNil(t, m.Counter)
	})
}

func TestUpdateCounter(t *testing.T) {
	d := newTestDriver(t)
	d.pipelineInfo.ID = "foo"

	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{Namespace: "test", Subsystem: "driver", Name: "counter"},
		[]string{"pipeline", "job"},
	)

	counterVecWithState := prometheus.NewCounterVec(
		prometheus.CounterOpts{Namespace: "test", Subsystem: "driver", Name: "counter_with_state"},
		[]string{"pipeline", "job", "state"},
	)

	// Passing a state to the stateless counter should error
	requireLogs(t, "expected 2 label values but got 3", func(logger logs.TaggedLogger) {
		d.updateCounter(counterVec, logger, "bar", func(c prometheus.Counter) {
			require.True(t, false, "should have errored")
		})
	})

	// updateCounter should pass a valid counter with the selected tags
	requireLogs(t, "", func(logger logs.TaggedLogger) {
		d.updateCounter(counterVec, logger, "", func(c prometheus.Counter) {
			c.Add(1)
		})
	})

	// Check that the counter was incremented
	requireCounter(t, counterVec, []string{"foo", "job-id"}, 1)

	// Not passing a state to the stateful counter should error
	requireLogs(t, "expected 3 label values but got 2", func(logger logs.TaggedLogger) {
		d.updateCounter(counterVecWithState, logger, "", func(c prometheus.Counter) {
			require.True(t, false, "should have errored")
		})
	})

	// updateCounter should pass a valid counter with the selected tags
	requireLogs(t, "", func(logger logs.TaggedLogger) {
		d.updateCounter(counterVecWithState, logger, "bar", func(c prometheus.Counter) {
			c.Add(1)
		})
	})

	// Check that the counter was incremented
	requireCounter(t, counterVecWithState, []string{"foo", "job-id", "bar"}, 1)
}

func TestUpdateHistogram(t *testing.T) {
	d := newTestDriver(t)
	d.pipelineInfo.ID = "foo"

	histogramVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "test", Subsystem: "driver", Name: "histogram",
			Buckets: prometheus.ExponentialBuckets(1.0, 2.0, 20),
		},
		[]string{"pipeline", "job"},
	)

	histogramVecWithState := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "test", Subsystem: "driver", Name: "histogram_with_state",
			Buckets: prometheus.ExponentialBuckets(1.0, 2.0, 20),
		},
		[]string{"pipeline", "job", "state"},
	)

	// Passing a state to the stateless histogram should error
	requireLogs(t, "expected 2 label values but got 3", func(logger logs.TaggedLogger) {
		d.updateHistogram(histogramVec, logger, "bar", func(h prometheus.Observer) {
			require.True(t, false, "should have errored")
		})
	})

	requireLogs(t, "", func(logger logs.TaggedLogger) {
		d.updateHistogram(histogramVec, logger, "", func(h prometheus.Observer) {
		})
	})

	// Check that the counter was incremented
	requireHistogram(t, histogramVec, []string{"foo", "job-id"}, 1)

	// Not passing a state to the stateful histogram should error
	requireLogs(t, "expected 3 label values but got 2", func(logger logs.TaggedLogger) {
		d.updateHistogram(histogramVecWithState, logger, "", func(h prometheus.Observer) {
			require.True(t, false, "should have errored")
		})
	})

	requireLogs(t, "", func(logger logs.TaggedLogger) {
		d.updateHistogram(histogramVecWithState, logger, "bar", func(h prometheus.Observer) {
		})
	})

	// Check that the counter was incremented
	requireHistogram(t, histogramVecWithState, []string{"foo", "job-id", "bar"}, 1)
}

/*
	ctx context.Context,
	data []*common.Input,
	inputTree *hashtree.Ordered,
	logger logs.TaggedLogger,
	cb func(*pps.ProcessStats) error,
*/

func provisionPipeline(d *driver) {
}

func TestWithData(t *testing.T) {
}

func TestWithDataCancel(t *testing.T) {
}

func TestWithDataGit(t *testing.T) {
}

func TestRunUserCode(t *testing.T) {
}

func TestRunUserErrorHandlingCode(t *testing.T) {
}

func TestUpdateJobState(t *testing.T) {
}

func TestDeleteJob(t *testing.T) {
}

/*
func lookupDockerUser(userArg string) (_ *user.User, retErr error) {
func lookupGroup(group string) (_ *user.Group, retErr error) {
func (d *driver) WithCtx(ctx context.Context) Driver {
func (d *driver) Jobs() col.Collection {
func (d *driver) Pipelines() col.Collection {
func (d *driver) Plans() col.Collection {
func (d *driver) Shards() col.Collection {
func (d *driver) Chunks(jobID string) col.Collection {
func (d *driver) Merges(jobID string) col.Collection {

func (d *driver) WithData(
func (d *driver) downloadData(
func (d *driver) downloadGitData(pachClient *client.APIClient, dir string, input *common.Input) error {
func (d *driver) linkData(inputs []*common.Input, dir string) error {
func (d *driver) unlinkData(inputs []*common.Input) error {

func (d *driver) RunUserCode(ctx context.Context, logger logs.TaggedLogger, environ []string, procStats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {
func (d *driver) RunUserErrorHandlingCode(ctx context.Context, logger logs.TaggedLogger, environ []string, procStats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {

func (d *driver) UpdateJobState(ctx context.Context, jobID string, statsCommit *pfs.Commit, state pps.JobState, reason string) error {
func (d *driver) DeleteJob(stm col.STM, jobPtr *pps.EtcdJobInfo) error {

func (d *driver) updateCounter(
func (d *driver) updateHistogram(
func (d *driver) reportUserCodeStats(logger logs.TaggedLogger) {
func (d *driver) reportDeferredUserCodeStats(err error, start time.Time, procStats *pps.ProcessStats, logger logs.TaggedLogger) {
func (d *driver) ReportUploadStats(start time.Time, procStats *pps.ProcessStats, logger logs.TaggedLogger) {
func (d *driver) reportDownloadSizeStats(downSize float64, logger logs.TaggedLogger) {
func (d *driver) reportDownloadTimeStats(start time.Time, procStats *pps.ProcessStats, logger logs.TaggedLogger) {
*/
