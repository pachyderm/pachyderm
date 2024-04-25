package logs_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	logservice "github.com/pachyderm/pachyderm/v2/src/server/logs"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/testloki"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

type testPublisher struct {
	responses []*logs.GetLogsResponse
}

func (tp *testPublisher) Publish(ctx context.Context, response *logs.GetLogsResponse) error {
	tp.responses = append(tp.responses, response)
	return nil
}

type errPublisher struct {
	count int
}

func (ep *errPublisher) Publish(context.Context, *logs.GetLogsResponse) error {
	ep.count++
	return errors.New("errPublisher cannot publish")
}

func TestGetLogsHint(t *testing.T) {
	var testData = map[string]struct {
		buildEntries func() []loki.Entry
		buildWant    func() []int
	}{
		"no logs to return": {
			buildEntries: func() []loki.Entry {
				return nil
			},
			buildWant: func() []int { return nil },
		},
		"logs": {
			buildEntries: func() []loki.Entry {
				var entries []loki.Entry
				for i := -99; i <= 0; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: time.Now().Add(time.Duration(i) * time.Second),
						Line:      fmt.Sprintf("%v", i),
					})
				}
				return entries
			},
			buildWant: func() []int {
				var want []int
				for i := -99; i <= 0; i++ {
					want = append(want, i)
				}
				return want
			},
		},
	}

	for name, testCase := range testData {
		t.Run(name, func(t *testing.T) {
			var (
				ctx      = pctx.TestContext(t)
				want     = testCase.buildWant()
				fakeLoki = httptest.NewServer(&lokiutil.FakeServer{
					Entries: testCase.buildEntries(),
				})
				ls = logservice.LogService{
					GetLokiClient: func() (*loki.Client, error) {
						return &loki.Client{Address: fakeLoki.URL}, nil
					},
				}
				publisher *testPublisher
			)
			defer fakeLoki.Close()

			// GetLogs without a hint request should not return hints.
			publisher = new(testPublisher)

			require.NoError(t, ls.GetLogs(ctx, &logs.GetLogsRequest{}, publisher), "GetLogs should succeed")
			for _, r := range publisher.responses {
				_, ok := r.ResponseType.(*logs.GetLogsResponse_PagingHint)
				require.False(t, ok, "paging hints should not be returned when unasked for")
			}
			require.Len(t, publisher.responses, len(want), "query with no date range should return all logs")

			// GetLogs with a hint request should return a hint.
			publisher = new(testPublisher)
			require.NoError(t, ls.GetLogs(ctx, &logs.GetLogsRequest{
				WantPagingHint: true,
			}, publisher), "GetLogs must succeed")
			require.True(t, len(publisher.responses) > 0, "there must be at least one response")
			var foundHint bool
			for _, resp := range publisher.responses {
				h, ok := resp.ResponseType.(*logs.GetLogsResponse_PagingHint)
				if ok && h != nil {
					foundHint = true
				}
			}
			require.True(t, foundHint, "paging hints should be returned when requested")

			// GetLogs with a hint request and a non-standard time filter should duplicate that duration.
			publisher = new(testPublisher)
			until := time.Now()
			from := until.Add(-1 * time.Second)
			require.NoError(t, ls.GetLogs(ctx, &logs.GetLogsRequest{
				WantPagingHint: true,
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From:  timestamppb.New(from),
						Until: timestamppb.New(until),
					},
				},
			}, publisher), "GetLogs with an explicit time filter must succeed")
			require.True(t, len(publisher.responses) > 0, "there must be at least one response")
			foundHint = false
			for _, resp := range publisher.responses {
				h, ok := resp.ResponseType.(*logs.GetLogsResponse_PagingHint)
				if !ok && h == nil {
					continue
				}
				foundHint = true
				for _, hint := range []*logs.GetLogsRequest{h.PagingHint.Older, h.PagingHint.Newer} {
					from := hint.Filter.TimeRange.From.AsTime()
					until := hint.Filter.TimeRange.Until.AsTime()
					require.Equal(t, time.Second, until.Sub(from), "explicit window is one hour, not %v (%v–%v)", until.Sub(from), from, until)
				}
			}
			require.True(t, foundHint, "paging hints should be returned when requested")

			publisher = new(testPublisher)
			require.NoError(t, ls.GetLogs(ctx, &logs.GetLogsRequest{
				WantPagingHint: true,
				Filter:         &logs.LogFilter{Limit: 100}}, publisher),
				"GetLogs with both a limit and a hint request should work")

			var badPublisher errPublisher
			err := ls.GetLogs(ctx, &logs.GetLogsRequest{}, &badPublisher)
			if badPublisher.count > 0 && !errors.Is(err, logservice.ErrPublish) {
				t.Error("GetLogs with a broken publisher must return an appropriate error)")
			}
			_ = want
		})
	}
}

func TestGetDatumLogs(t *testing.T) {
	var (
		ctx             = pctx.TestContext(t)
		foundQuery      bool
		datumMiddleware = func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				q := req.URL.Query()
				if q := q.Get("query"); q != "" {
					if strings.Contains(q, `"12d0b112f8deea684c4530693545901608bfb088d564d3c68dddaf2a02d446f5"`) {
						foundQuery = true
					}
				}
				next.ServeHTTP(rw, req)
			})
		}
		fakeLoki = httptest.NewServer(datumMiddleware(&lokiutil.FakeServer{}))
		ls       = logservice.LogService{
			GetLokiClient: func() (*loki.Client, error) {
				return &loki.Client{Address: fakeLoki.URL}, nil
			},
		}
		publisher *testPublisher
	)
	defer fakeLoki.Close()
	publisher = new(testPublisher)
	require.NoError(t, ls.GetLogs(ctx, &logs.GetLogsRequest{
		Query: &logs.LogQuery{
			QueryType: &logs.LogQuery_User{
				User: &logs.UserLogQuery{
					UserType: &logs.UserLogQuery_Datum{
						Datum: "12d0b112f8deea684c4530693545901608bfb088d564d3c68dddaf2a02d446f5",
					},
				},
			},
		},
	}, publisher), "GetLogs should succeed")
	require.True(t, foundQuery, "datum LogQL query should be found")

}

func TestGetProjectLogs(t *testing.T) {
	var (
		ctx             = pctx.TestContext(t)
		foundQuery      bool
		projectName     = testutil.UniqueString("pipelineProject")
		datumMiddleware = func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				q := req.URL.Query()
				if q := q.Get("query"); q != "" {
					if strings.Contains(q, fmt.Sprintf(`pipelineProject="%s"`, projectName)) {
						foundQuery = true
					}
				}
				next.ServeHTTP(rw, req)
			})
		}
		fakeLoki = httptest.NewServer(datumMiddleware(&lokiutil.FakeServer{}))
		ls       = logservice.LogService{
			GetLokiClient: func() (*loki.Client, error) {
				return &loki.Client{Address: fakeLoki.URL}, nil
			},
		}
		publisher *testPublisher
	)
	defer fakeLoki.Close()
	publisher = new(testPublisher)
	require.NoError(t, ls.GetLogs(ctx, &logs.GetLogsRequest{
		Query: &logs.LogQuery{
			QueryType: &logs.LogQuery_User{
				User: &logs.UserLogQuery{
					UserType: &logs.UserLogQuery_Project{
						Project: projectName,
					},
				},
			},
		},
	}, publisher), "GetLogs should succeed")
	require.True(t, foundQuery, "project LogQL query should be found")

}

func TestPipelineLogs(t *testing.T) {
	ctx := pctx.TestContext(t)
	aloki, err := testloki.New(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("new test loki: %v", err)
	}
	t.Cleanup(func() {
		if err := aloki.Close(); err != nil {
			t.Fatalf("clean up loki: %v", err)
		}
	})
	labels := map[string]string{
		"app":             "pipeline",
		"component":       "worker",
		"container":       "user",
		"pipelineName":    "edges",
		"pipelineProject": "default",
		"pipelineVersion": "1",
		"suite":           "pachyderm",
	}
	hash, err := base64.StdEncoding.DecodeString("Vp/ZEXxcM96lYfLUnnuaFECJ1j4tuvla7TsY6XGF7qU=")
	if err != nil {
		t.Fatalf("base64 error: %v", err)
	}
	messageTexts := []string{
		"/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment. 1",
		"  warnings.warn('Matplotlib is building the font cache using fc-list. This may take a moment.') 2",
		"/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment. 3",
		"  warnings.warn('Matplotlib is building the font cache using fc-list. This may take a moment.') 4",
	}
	now := time.Now().Add(-(time.Duration(len(messageTexts)) + 1) * time.Second)
	ts := []time.Time{
		now,
		now.Add(time.Second),
		now.Add(2 * time.Second),
		now.Add(3 * time.Second),
	}
	pipelineLogs := []*testloki.Log{
		{
			Time: now,
			Labels: map[string]string{
				"app":   "pachd",
				"suite": "pachyderm",
			},
			Message: `{"severity":"info","time":"2024-04-16T21:10:36.717965399Z","caller":"pachd/main.go:40","message":"pachd: starting","mode":"full"}`,
		},
		{
			Time:    ts[0],
			Labels:  labels,
			Message: `{"severity":"info","ts":"2024-04-16T21:10:37.234151650Z","message":"` + messageTexts[0] + `","workerId":"default-edges-v1-8sx6n","projectName":"default","pipelineName":"edges","jobId":"c4cae897bc914bd4bdb6262db038ff15","data":[{"path":"/liberty.jpg","hash":"Vp/ZEXxcM96lYfLUnnuaFECJ1j4tuvla7TsY6XGF7qU="}],"datumId":"cef4a52be60465b328ea783037b6c46531ad7cc9f4e190f9c0e548f473cd1fd1","user":true,"stream":"stderr"}`,
		},
		{
			Time:    ts[1],
			Labels:  labels,
			Message: `{"severity":"info","ts":"2024-04-16T21:10:37.234183251Z","message":"` + messageTexts[1] + `","workerId":"default-edges-v1-8sx6n","projectName":"default","pipelineName":"edges","jobId":"c4cae897bc914bd4bdb6262db038ff15","data":[{"path":"/liberty.jpg","hash":"Vp/ZEXxcM96lYfLUnnuaFECJ1j4tuvla7TsY6XGF7qU="}],"datumId":"cef4a52be60465b328ea783037b6c46531ad7cc9f4e190f9c0e548f473cd1fd1","user":true,"stream":"stderr"}`,
		},
		{
			Time:    ts[2],
			Labels:  labels,
			Message: `{"severity":"info","ts":"2024-04-16T21:10:37.234186851Z","message":"` + messageTexts[2] + `","workerId":"default-edges-v1-8sx6n","projectName":"default","pipelineName":"edges","jobId":"c4cae897bc914bd4bdb6262db038ff15","data":[{"path":"/liberty.jpg","hash":"Vp/ZEXxcM96lYfLUnnuaFECJ1j4tuvla7TsY6XGF7qU="}],"datumId":"cef4a52be60465b328ea783037b6c46531ad7cc9f4e190f9c0e548f473cd1fd1","user":true,"stream":"stderr"}`,
		},
		{
			Time:    ts[3],
			Labels:  labels,
			Message: `{"severity":"info","ts":"2024-04-16T21:10:37.234188451Z","message":"` + messageTexts[3] + `","workerId":"default-edges-v1-8sx6n","projectName":"default","pipelineName":"edges","jobId":"c4cae897bc914bd4bdb6262db038ff15","data":[{"path":"/liberty.jpg","hash":"Vp/ZEXxcM96lYfLUnnuaFECJ1j4tuvla7TsY6XGF7qU="}],"datumId":"cef4a52be60465b328ea783037b6c46531ad7cc9f4e190f9c0e548f473cd1fd1","user":true,"stream":"stderr"}`,
		},
	}

	wants := []*logs.GetLogsResponse{
		{
			ResponseType: &logs.GetLogsResponse_Log{
				Log: &logs.LogMessage{
					NativeTimestamp: timestamppb.New(time.Date(2024, 04, 16, 21, 10, 37, 234151650, time.UTC)),
					Verbatim: &logs.VerbatimLogMessage{
						Line:      []byte(pipelineLogs[1].Message),
						Timestamp: timestamppb.New(pipelineLogs[1].Time),
					},
				},
			},
		},
		{
			ResponseType: &logs.GetLogsResponse_Log{
				Log: &logs.LogMessage{
					NativeTimestamp: timestamppb.New(time.Date(2024, 04, 16, 21, 10, 37, 234183251, time.UTC)),
					Verbatim: &logs.VerbatimLogMessage{
						Line:      []byte(pipelineLogs[2].Message),
						Timestamp: timestamppb.New(pipelineLogs[2].Time),
					},
				},
			},
		},
		{
			ResponseType: &logs.GetLogsResponse_Log{
				Log: &logs.LogMessage{
					NativeTimestamp: timestamppb.New(time.Date(2024, 04, 16, 21, 10, 37, 234186851, time.UTC)),
					Verbatim: &logs.VerbatimLogMessage{
						Line:      []byte(pipelineLogs[3].Message),
						Timestamp: timestamppb.New(pipelineLogs[3].Time),
					},
				},
			},
		},
		{
			ResponseType: &logs.GetLogsResponse_Log{
				Log: &logs.LogMessage{
					NativeTimestamp: timestamppb.New(time.Date(2024, 04, 16, 21, 10, 37, 234188451, time.UTC)),
					Verbatim: &logs.VerbatimLogMessage{
						Line:      []byte(pipelineLogs[4].Message),
						Timestamp: timestamppb.New(pipelineLogs[4].Time),
					},
				},
			},
		},
	}

	for i := range wants {
		log := pipelineLogs[i+1] // the 0th piece of testdata should not be returned.
		object := new(structpb.Struct)
		if err := object.UnmarshalJSON([]byte(log.Message)); err != nil {
			t.Fatalf("failed to unmarshal json into protobuf Struct, %q, %q", err, log.Message)
		}
		wants[i].GetLog().Object = object
		wants[i].GetLog().PpsLogMessage = &pps.LogMessage{
			ProjectName:  "default",
			PipelineName: "edges",
			JobId:        "c4cae897bc914bd4bdb6262db038ff15",
			WorkerId:     "default-edges-v1-8sx6n",
			DatumId:      "cef4a52be60465b328ea783037b6c46531ad7cc9f4e190f9c0e548f473cd1fd1",
			Data:         []*pps.InputFile{{Path: "/liberty.jpg", Hash: hash}},
			User:         true,
			Ts:           wants[i].GetLog().NativeTimestamp,
			Message:      messageTexts[i],
		}
	}
	for _, log := range pipelineLogs {
		if err := aloki.AddLog(ctx, log); err != nil {
			t.Fatalf("add log: %v", err)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ls := logservice.LogService{
		GetLokiClient: func() (*loki.Client, error) {
			return aloki.Client, nil
		},
	}
	publisher := new(testPublisher)
	require.NoError(t, ls.GetLogs(ctx, &logs.GetLogsRequest{
		Query: &logs.LogQuery{
			QueryType: &logs.LogQuery_User{
				User: &logs.UserLogQuery{
					UserType: &logs.UserLogQuery_Pipeline{
						Pipeline: &logs.PipelineLogQuery{
							Project:  "default",
							Pipeline: "edges",
						},
					},
				},
			},
		},
	}, publisher), "GetLogs should succeed")
	require.NoDiff(t, wants, publisher.responses, []cmp.Option{protocmp.Transform()})
}
