package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/otiai10/copy"
	"github.com/prometheus/client_golang/prometheus"
	prometheus_proto "github.com/prometheus/client_model/go"
	"gopkg.in/go-playground/webhooks.v5/github"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

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

type testEnv struct {
	tu.Env
	mockKube *MockKubeWrapper
	driver   *driver
}

func withTestEnv(cb func(*testEnv)) error {
	return tu.WithEnv(func(baseEnv *tu.Env) (err error) {
		env := &testEnv{Env: *baseEnv}

		// Mock out the enterprise.GetState call that happens during driver construction
		env.MockPachd.Enterprise.GetState.Use(func(context.Context, *enterprise.GetStateRequest) (*enterprise.GetStateResponse, error) {
			return &enterprise.GetStateResponse{State: enterprise.State_NONE}, nil
		})

		env.mockKube = NewMockKubeWrapper().(*MockKubeWrapper)

		var d Driver
		d, err = NewDriver(
			testPipelineInfo,
			env.PachClient,
			env.mockKube,
			env.EtcdClient,
			tu.UniqueString("driverTest"),
		)
		d = d.WithCtx(env.Context)
		env.driver = d.(*driver)
		if err != nil {
			return err
		}

		env.driver.inputDir = filepath.Clean(path.Join(env.Directory, "pfs"))

		cb(env)

		return nil
	})
}

// Note: this function only exists for tests, the real system uses a fifo for
// this (which does not exist in the normal filesystem on Windows)
func createSpoutFifo(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}

// os.Symlink requires additional privileges on windows, so just copy the files instead
func (d *driver) linkData(inputs []*common.Input, dir string) error {
	// Make sure that previously symlinked outputs are removed.
	if err := d.unlinkData(inputs); err != nil {
		return err
	}
	for _, input := range inputs {
		src := filepath.Join(dir, input.Name)
		dst := filepath.Join(d.inputDir, input.Name)
		if err := copy.Copy(src, dst); err != nil {
			duration, err := time.ParseDuration("1000s")
			time.Sleep(duration)
			return err
		}
	}
	return copy.Copy(filepath.Join(dir, "out"), filepath.Join(d.inputDir, "out"))
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

func requireHistogram(t *testing.T, histogram *prometheus.HistogramVec, labels []string, value uint64) {
	requireMetric(t, histogram, labels, func(m prometheus_proto.Metric) {
		require.NotNil(t, m.Histogram)
		require.Equal(t, value, *m.Histogram.SampleCount)
	})
}

func TestUpdateCounter(t *testing.T) {
	err := withTestEnv(func(env *testEnv) {
		env.driver.pipelineInfo.ID = "foo"

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
			env.driver.updateCounter(counterVec, logger, "bar", func(c prometheus.Counter) {
				require.True(t, false, "should have errored")
			})
		})

		// updateCounter should pass a valid counter with the selected tags
		requireLogs(t, "", func(logger logs.TaggedLogger) {
			env.driver.updateCounter(counterVec, logger, "", func(c prometheus.Counter) {
				c.Add(1)
			})
		})

		// Check that the counter was incremented
		requireCounter(t, counterVec, []string{"foo", "job-id"}, 1)

		// Not passing a state to the stateful counter should error
		requireLogs(t, "expected 3 label values but got 2", func(logger logs.TaggedLogger) {
			env.driver.updateCounter(counterVecWithState, logger, "", func(c prometheus.Counter) {
				require.True(t, false, "should have errored")
			})
		})

		// updateCounter should pass a valid counter with the selected tags
		requireLogs(t, "", func(logger logs.TaggedLogger) {
			env.driver.updateCounter(counterVecWithState, logger, "bar", func(c prometheus.Counter) {
				c.Add(1)
			})
		})

		// Check that the counter was incremented
		requireCounter(t, counterVecWithState, []string{"foo", "job-id", "bar"}, 1)
	})
	require.NoError(t, err)
}

func TestUpdateHistogram(t *testing.T) {
	err := withTestEnv(func(env *testEnv) {
		env.driver.pipelineInfo.ID = "foo"

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
			env.driver.updateHistogram(histogramVec, logger, "bar", func(h prometheus.Observer) {
				require.True(t, false, "should have errored")
			})
		})

		requireLogs(t, "", func(logger logs.TaggedLogger) {
			env.driver.updateHistogram(histogramVec, logger, "", func(h prometheus.Observer) {
				h.Observe(0)
			})
		})

		// Check that the counter was incremented
		requireHistogram(t, histogramVec, []string{"foo", "job-id"}, 1)

		// Not passing a state to the stateful histogram should error
		requireLogs(t, "expected 3 label values but got 2", func(logger logs.TaggedLogger) {
			env.driver.updateHistogram(histogramVecWithState, logger, "", func(h prometheus.Observer) {
				require.True(t, false, "should have errored")
			})
		})

		requireLogs(t, "", func(logger logs.TaggedLogger) {
			env.driver.updateHistogram(histogramVecWithState, logger, "bar", func(h prometheus.Observer) {
				h.Observe(0)
			})
		})

		// Check that the counter was incremented
		requireHistogram(t, histogramVecWithState, []string{"foo", "job-id", "bar"}, 1)
	})
	require.NoError(t, err)
}

type inputData struct {
	path     string
	contents string
	regex    string
	found    bool
}

func newInputDataRegex(path string, regex string) *inputData {
	return &inputData{path: filepath.Clean(path), regex: regex}
}

func newInputData(path string, contents string) *inputData {
	return &inputData{path: filepath.Clean(path), contents: contents}
}

func requireInputContents(t *testing.T, env *testEnv, data []*inputData) {
	checkFile := func(fullPath string, relPath string) {
		for _, checkData := range data {
			if checkData.path == relPath {
				contents, err := ioutil.ReadFile(fullPath)
				require.NoError(t, err)
				if checkData.regex != "" {
					require.Matches(t, checkData.regex, string(contents), "Incorrect contents for input file: %s", relPath)
				} else {
					require.Equal(t, checkData.contents, string(contents), "Incorrect contents for input file: %s", relPath)
				}
				checkData.found = true
				return
			}
		}
		require.True(t, false, "Unexpected input file found: %s", relPath)
	}

	err := filepath.Walk(env.driver.inputDir, func(path string, info os.FileInfo, err error) error {
		if info.Name() == ".scratch" || info.Name() == ".git" {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			path = filepath.Clean(path)
			relPath := strings.TrimLeft(strings.TrimPrefix(path, env.driver.inputDir), "/\\")
			checkFile(path, relPath)
		}
		return nil
	})
	require.NoError(t, err)

	for _, checkData := range data {
		require.True(t, checkData.found, "Expected input file not found: %s", checkData.path)
	}
}

func TestWithDataEmpty(t *testing.T) {
	err := withTestEnv(func(env *testEnv) {
		requireLogs(t, "finished downloading data", func(logger logs.TaggedLogger) {
			_, err := env.driver.WithData(
				[]*common.Input{},
				nil,
				logger,
				func(stats *pps.ProcessStats) error {
					requireInputContents(t, env, []*inputData{})
					return nil
				},
			)
			require.NoError(t, err)
			requireInputContents(t, env, []*inputData{})
		})
	})
	require.NoError(t, err)
}

func TestWithDataSpout(t *testing.T) {
	err := withTestEnv(func(env *testEnv) {
		env.driver.pipelineInfo.Spout = &pps.Spout{}
		requireLogs(t, "finished downloading data", func(logger logs.TaggedLogger) {
			_, err := env.driver.WithData(
				[]*common.Input{},
				nil,
				logger,
				func(stats *pps.ProcessStats) error {
					// A spout pipeline should have created a 'pfs/out` fifo for the user
					// code to write to
					requireInputContents(t, env, []*inputData{newInputData("out", "")})
					return nil
				},
			)
			require.NoError(t, err)
			requireInputContents(t, env, []*inputData{})
		})
	})
	require.NoError(t, err)
}

// Shitty helper function to create possibly-not-malformed input structures
func newInput(repo string, path string) *common.Input {
	return &common.Input{
		FileInfo: &pfs.FileInfo{
			File: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: repo,
					},
					ID: "commit-id-string",
				},
				Path: path,
			},
			FileType: pfs.FileType_FILE,
		},
		Name:   repo,
		Branch: "master",
	}
}

func TestWithDataCancel(t *testing.T) {
	err := withTestEnv(func(env *testEnv) {
		requireLogs(t, "errored downloading data.*context canceled", func(logger logs.TaggedLogger) {
			ctx, cancel := context.WithCancel(env.Context)
			driver := env.driver.WithCtx(ctx)

			// Cancel the context during the download
			env.MockPachd.PFS.WalkFile.Use(func(req *pfs.WalkFileRequest, serv pfs.API_WalkFileServer) error {
				cancel()
				<-serv.Context().Done()
				return fmt.Errorf("WalkFile canceled")
			})

			_, err := driver.WithData(
				[]*common.Input{newInput("repo", "input.txt")},
				nil,
				logger,
				func(stats *pps.ProcessStats) error {
					require.True(t, false, "Should have been canceled before the callback")
					cancel()
					return nil
				},
			)
			require.YesError(t, err, "WithData call should have been canceled")
			requireInputContents(t, env, []*inputData{})
		})
	})
	require.NoError(t, err)
}

// Check that the driver will download the requested inputs, put them in place
// during WithData, and clean them up after running the inner function.
func TestWithDataDownload(t *testing.T) {
	err := withTestEnv(func(env *testEnv) {
		requireLogs(t, "finished downloading data.*inner function", func(logger logs.TaggedLogger) {
			// Mock out the calls that will be used to download the data
			env.MockPachd.PFS.WalkFile.Use(func(req *pfs.WalkFileRequest, serv pfs.API_WalkFileServer) error {
				return serv.Send(&pfs.FileInfo{
					File:     req.File,
					FileType: pfs.FileType_FILE,
				})
			})

			env.MockPachd.PFS.GetFile.Use(func(req *pfs.GetFileRequest, serv pfs.API_GetFileServer) error {
				return serv.Send(&types.BytesValue{Value: []byte(fmt.Sprintf("%s-data", req.File.Commit.Repo.Name))})
			})

			_, err := env.driver.WithData(
				[]*common.Input{newInput("repoA", "input.txt"), newInput("repoB", "input.md")},
				nil,
				logger,
				func(stats *pps.ProcessStats) error {
					requireInputContents(t, env, []*inputData{
						newInputData("repoA/input.txt", "repoA-data"),
						newInputData("repoB/input.md", "repoB-data"),
					})
					logger.Logf("inner function")
					return nil
				},
			)
			require.NoError(t, err)
			requireInputContents(t, env, []*inputData{})
		})
	})
	require.NoError(t, err)
}

// Create several files and directories inside WithData and verify that they are
// cleaned up after WithData returns.
func TestWithDataCleanup(t *testing.T) {
	err := withTestEnv(func(env *testEnv) {
		create := func(relPath string) {
			fullPath := path.Join(env.driver.inputDir, relPath)
			require.NoError(t, os.MkdirAll(path.Dir(fullPath), 0777))
			file, err := os.Create(fullPath)
			require.NoError(t, err)
			require.NoError(t, file.Close())
		}

		requireLogs(t, "finished downloading data.*inner function", func(logger logs.TaggedLogger) {
			_, err := env.driver.WithData(
				[]*common.Input{},
				nil,
				logger,
				func(stats *pps.ProcessStats) error {
					requireInputContents(t, env, []*inputData{})
					logger.Logf("inner function")

					create("c")
					create("out/1")
					create("out/2/a")
					create("out/2/b")
					create("out/2/3/c")
					create("foo/barbaz")
					create("foo/bar/baz")
					create("floop/blarp/blazj/etc")

					return nil
				},
			)
			require.NoError(t, err)
			requireInputContents(t, env, []*inputData{})
		})
	})
	require.NoError(t, err)
}

func newGitInput(repo string, url string) *common.Input {
	return &common.Input{
		FileInfo: &pfs.FileInfo{
			File: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: repo,
					},
					ID: "commit-id-string",
				},
				Path: "commit.json",
			},
			FileType: pfs.FileType_FILE,
		},
		GitURL: url,
		Name:   repo,
	}
}

func TestWithDataGit(t *testing.T) {
	err := withTestEnv(func(env *testEnv) {
		requireLogs(t, "finished downloading data", func(logger logs.TaggedLogger) {
			var getFileReq *pfs.GetFileRequest
			env.MockPachd.PFS.GetFile.Use(func(req *pfs.GetFileRequest, serv pfs.API_GetFileServer) (retErr error) {
				getFileReq = req

				payload := &github.PushPayload{
					Ref:   "refs/heads/master",
					After: "foobar",
				}
				payload.Repository.CloneURL = "https://github.com/pachyderm/test-artifacts.git"
				jsonBytes, err := json.Marshal(payload)
				if err != nil {
					return err
				}

				return serv.Send(&types.BytesValue{Value: jsonBytes})
			})
			_, err := env.driver.WithData(
				[]*common.Input{
					newGitInput("artifacts", "https://github.com/pachyderm/test-artifacts.git"),
				},
				nil,
				logger,
				func(stats *pps.ProcessStats) error {
					requireInputContents(t, env, []*inputData{newInputDataRegex("artifacts/readme.md", "Test Artifacts")})
					return nil
				},
			)
			require.NoError(t, err)
			require.NotNil(t, getFileReq)
			require.Equal(t, getFileReq.File, client.NewFile("artifacts", "commit-id-string", "commit.json"))
			requireInputContents(t, env, []*inputData{})
		})
	})
	require.NoError(t, err)
}

func TestWithDataGitError(t *testing.T) {
}

func TestWithDataStats(t *testing.T) {
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
