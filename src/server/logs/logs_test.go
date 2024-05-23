package logs_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	logservice "github.com/pachyderm/pachyderm/v2/src/server/logs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/testloki"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

type testPublisher struct {
	responses []*logs.GetLogsResponse
}

func (tp *testPublisher) Publish(ctx context.Context, response *logs.GetLogsResponse) error {
	tp.responses = append(tp.responses, response)
	return nil
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

type hintCase struct {
	olderFrom, olderUntil time.Duration
	olderOffset           uint
	newerFrom, newerUntil time.Duration
	newerOffset           uint
}
type testCase struct {
	logs        []time.Duration // offsets from time.Now()
	limit       uint
	from, until time.Duration // ditto
	offset      uint
	want        []time.Duration
	wantHint    *hintCase
}

func TestGetLogs_missingFromUntil(t *testing.T) {
	var (
		ctx        = pctx.TestContext(t)
		now        = time.Now()
		aloki, err = testloki.New(ctx, t.TempDir())
		ls         = logservice.LogService{
			GetLokiClient: func() (*loki.Client, error) {
				return aloki.Client, nil
			},
		}

		publisher = new(testPublisher)
		req       = &logs.GetLogsRequest{
			WantPagingHint: true,
			Filter:         &logs.LogFilter{},
			Query: &logs.LogQuery{
				QueryType: &logs.LogQuery_Admin{
					Admin: &logs.AdminLogQuery{
						AdminType: &logs.AdminLogQuery_Logql{
							Logql: `{app="testpach"}`,
						},
					},
				},
			},
		}
		testLogs                 = []time.Duration{-700*time.Hour - time.Second, -1 * time.Hour, -1 * time.Minute, -1 * time.Second, time.Minute, time.Hour}
		want                     = []time.Duration{-1 * time.Hour, -1 * time.Minute, -1 * time.Second}
		olderFrom                = -700 * time.Hour
		olderUntil time.Duration = -1400 * time.Hour
		newerFrom  time.Duration = 0
		newerUntil               = 700 * time.Hour
	)
	if err != nil {
		t.Fatalf("new test loki: %v", err)
	}
	t.Cleanup(func() {
		if err := aloki.Close(); err != nil {
			t.Fatalf("clean up loki: %v", err)
		}
	})
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for i, offset := range testLogs {
		var (
			timestamp = now.Add(offset)
			pachLog   = &pps.LogMessage{
				Ts:      timestamppb.New(timestamp),
				Message: fmt.Sprintf("log %d", i),
			}
			b   []byte
			log = &testloki.Log{
				Time: timestamp,
				Labels: map[string]string{
					"app": "testpach",
				},
			}
			err error
		)
		if b, err = protojson.Marshal(pachLog); err != nil {
			t.Fatal(err)
		}
		log.Message = string(b)
		if err = aloki.AddLog(ctx, log); err != nil {
			t.Fatal(err)
		}
	}
	require.NoError(t, ls.GetLogs(ctx, req, publisher), "GetLogs should succeed")
	var hints []*logs.PagingHint
	for _, resp := range publisher.responses {
		if hint := resp.GetPagingHint(); hint != nil {
			hints = append(hints, hint)
		}
	}
	if len(publisher.responses)-len(hints) != len(want) {
		for i, resp := range publisher.responses {
			t.Logf("resp %d: %v", i, resp.GetLog())
		}
		t.Fatalf("got %d responses; want %d", len(publisher.responses)-len(hints), len(want))
	}
	if len(hints) == 0 {
		t.Errorf("wanted hints; got none")
	}
	for i, hint := range hints {
		if hint == nil || (hint.Older == nil && hint.Newer == nil) {
			t.Errorf("hint %d is empty", i)
		}
		// Since we cannot know the server-side time, and the default is
		// set server side, check that there is a consistent delta.
		//
		// NOTE: this assumes that a single hint contains both older and
		// newer hints; it won’t handle things properly if they are
		// split up.  Doing that is trickier than it is probably worth
		// right now.
		var delta time.Duration
		if hint.Older != nil {
			if got, want := hint.Older.Filter.TimeRange.From.AsTime(), now.Add(olderFrom); !got.Equal(want) {
				delta = want.Sub(got)
			}
			if got, want := hint.Older.Filter.TimeRange.Until.AsTime().Add(delta), now.Add(olderUntil); !got.Equal(want) {
				t.Errorf("wanted older hint until = %v; got %v (Δ %v)", want, got, want.Sub(got))
			}
		}
		if hint.Newer != nil {
			if got, want := hint.Newer.Filter.TimeRange.From.AsTime().Add(delta), now.Add(newerFrom); !got.Equal(want) {
				t.Errorf("wanted newer hint from = %v; got %v (Δ %v)", want, got, want.Sub(got))
			}
			if got, want := hint.Newer.Filter.TimeRange.Until.AsTime().Add(delta), now.Add(newerUntil); !got.Equal(want) {
				t.Errorf("wanted newer hint until = %v; got %v (Δ %v)", want, got, want.Sub(got))
			}

		}
	}
	for i := range want {
		if want, got := now.Add(want[i]), publisher.responses[i].GetLog().GetVerbatim().GetTimestamp().AsTime(); !want.Equal(got) {
			t.Errorf("expected item %d to be %v; got %v", i, want, got)
		}
	}

}

func TestGetLogs_offset(t *testing.T) {
	var testCases = map[string]testCase{
		"no logs at all should return no logs": {
			logs:  []time.Duration{},
			limit: 0,
			from:  time.Second,
			until: time.Second * 12,
			want:  nil,
		},
		"no logs in window should return no logs": {
			logs:  []time.Duration{0},
			limit: 0,
			from:  time.Second,
			until: time.Second * 12,
			want:  nil,
		},
		"a log in window should return that log": {
			logs:  []time.Duration{time.Second * 2},
			limit: 0,
			from:  time.Second,
			until: time.Second * 12,
			want:  []time.Duration{time.Second * 2},
		},
		"logs outside a window should not be returned": {
			logs:  []time.Duration{time.Nanosecond, time.Second * 3 / 2, time.Second * 2, time.Second * 14},
			limit: 0,
			from:  time.Second,
			until: time.Second * 12,
			want:  []time.Duration{time.Second * 3 / 2, time.Second * 2},
		},
		"from is inclusive; until is exclusive": {
			logs:  []time.Duration{time.Nanosecond, time.Second, time.Second * 3 / 2, time.Second * 2, time.Second * 12, time.Second * 14},
			limit: 0,
			from:  time.Second,
			until: time.Second * 12,
			want:  []time.Duration{time.Second, time.Second * 3 / 2, time.Second * 2},
		},
		"from after until is backwards": {
			logs:  []time.Duration{time.Nanosecond, time.Second, time.Second * 3 / 2, time.Second * 2, time.Second * 12, time.Second * 14},
			limit: 0,
			from:  time.Second * 12,
			until: time.Second,
			want:  []time.Duration{time.Second * 2, time.Second * 3 / 2, time.Second},
		},
		// If there is a log in the window without a limit, then the
		// older window ends at the request from and the newer window
		// starts at the request until, with no offset.
		"hinting works": {
			logs:  []time.Duration{time.Second * 2},
			limit: 0,
			from:  time.Second,
			until: time.Second * 12,
			want:  []time.Duration{time.Second * 2},
			wantHint: &hintCase{
				olderFrom:  time.Second,
				olderUntil: time.Second * -10,
				newerFrom:  time.Second * 12,
				newerUntil: time.Second * 23,
			},
		},
		// If there is a log in the window with a limit, then the older
		// window still ends at the request from, but the newer window
		// starts at the same time as the last log, with an offset of
		// one.
		"hinting works with limit": {
			logs:  []time.Duration{1, time.Second * 2, time.Second * 3},
			limit: 1,
			from:  time.Second,
			until: time.Second * 12,
			want:  []time.Duration{time.Second * 2},
			wantHint: &hintCase{
				olderFrom:   time.Second,
				olderUntil:  time.Second * -10,
				newerFrom:   time.Second * 2,
				newerUntil:  time.Second * 13,
				newerOffset: 1,
			},
		},
		// check that the older hint above works
		"returned older hint works": {
			logs:  []time.Duration{1, time.Second * 2, time.Second * 3},
			limit: 1,
			from:  time.Second,
			until: time.Second * -10,
			want:  []time.Duration{1},
		},
		// check that the newer hint above works
		"returned hint works": {
			logs:   []time.Duration{1, time.Second * 2, time.Second * 3},
			limit:  1,
			from:   time.Second * 2,
			until:  time.Second * 13,
			offset: 1,
			want:   []time.Duration{time.Second * 3},
		},
		// If there is a run _within_ a window without a limit, then the
		// newer hint should not have an offset.
		"hinting works after run": {
			logs:  []time.Duration{time.Second * 3 / 2, time.Second * 2, time.Second * 2},
			limit: 0,
			from:  time.Second,
			until: time.Second * 12,
			want:  []time.Duration{time.Second * 3 / 2, time.Second * 2, time.Second * 2},
			wantHint: &hintCase{
				olderFrom:  time.Second,
				olderUntil: time.Second * -10,
				newerFrom:  time.Second * 12,
				newerUntil: time.Second * 23,
			},
		},
		// If there is a log in the window WITH a limit, then the older
		// window ends at the request from and the newer window.
		"limit works": {
			logs:  []time.Duration{time.Second * 2, time.Second * 3, time.Second * 4},
			limit: 2,
			from:  time.Second,
			until: time.Second * 12,
			want:  []time.Duration{time.Second * 2, time.Second * 3},
		},
		"limit works with hint": {
			logs:  []time.Duration{time.Second * 2, time.Second * 3, time.Second * 4},
			limit: 2,
			from:  time.Second,
			until: time.Second * 12,
			want:  []time.Duration{time.Second * 2, time.Second * 3},
			wantHint: &hintCase{
				olderFrom:   time.Second,
				olderUntil:  time.Second * -10,
				newerFrom:   time.Second * 3,
				newerUntil:  time.Second * 14,
				newerOffset: 1,
			},
		},
		"offset works": {
			logs:  []time.Duration{time.Second * 2, time.Second * 3, time.Second * 3},
			limit: 2,
			from:  time.Second,
			until: time.Second * 12,
			want:  []time.Duration{time.Second * 2, time.Second * 3},
			wantHint: &hintCase{
				olderFrom:   time.Second,
				olderUntil:  time.Second * -10,
				newerFrom:   time.Second * 3,
				newerUntil:  time.Second * 14,
				newerOffset: 1,
			},
		},
		"backwards older hint is correct": {
			logs:  []time.Duration{time.Second * -3, time.Second * -2, time.Second * -1, 0, time.Second, time.Second * 2},
			limit: 0,
			from:  time.Second * 10,
			until: time.Second * -1,
			want:  []time.Duration{time.Second * 2, time.Second, 0, time.Second * -1},
			wantHint: &hintCase{
				olderFrom:  time.Second * -1,
				olderUntil: time.Second * -12,
				newerFrom:  time.Second * 10,
				newerUntil: time.Second * 21,
			},
		},
		"backwards older hint works": {

			logs:  []time.Duration{time.Second * -3, time.Second * -2, time.Second * -1, 0, time.Second, time.Second * 2},
			limit: 0,
			from:  time.Second * -1,
			until: time.Second * -12,
			want:  []time.Duration{time.Second * -2, time.Second * -3},
			wantHint: &hintCase{
				olderFrom:  time.Second * -12,
				olderUntil: time.Second * -23,
				newerFrom:  time.Second * -1,
				newerUntil: time.Second * 10,
			},
		},
		"backwards older hint with a limit is correct": {
			logs:  []time.Duration{time.Second * -3, time.Second * -2, time.Second * -1, 0, time.Second, time.Second * 2},
			limit: 2,
			from:  time.Second * 10,
			until: time.Second * -1,
			want:  []time.Duration{time.Second * 2, time.Second},
			wantHint: &hintCase{
				olderFrom:  time.Second,
				olderUntil: time.Second * -10,
				newerFrom:  time.Second * 10,
				newerUntil: time.Second * 21,
			},
		},
		"backwards older hint with a limit works": {
			logs:  []time.Duration{time.Second * -3, time.Second * -2, time.Second * -1, 0, time.Second, time.Second * 2},
			limit: 2,
			from:  time.Second,
			until: time.Second * -10,
			want:  []time.Duration{0, time.Second * -1},
			wantHint: &hintCase{
				olderFrom:  time.Second * -1,
				olderUntil: time.Second * -12,
				newerFrom:  time.Second,
				newerUntil: time.Second * 12,
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var (
				ctx        = pctx.TestContext(t)
				now        = time.Now()
				aloki, err = testloki.New(ctx, t.TempDir())
				ls         = logservice.LogService{
					GetLokiClient: func() (*loki.Client, error) {
						return aloki.Client, nil
					},
				}

				publisher = new(testPublisher)
				req       = &logs.GetLogsRequest{
					Filter: &logs.LogFilter{
						Limit: uint64(tc.limit),
						TimeRange: &logs.TimeRangeLogFilter{
							From:   timestamppb.New(now.Add(tc.from)),
							Until:  timestamppb.New(now.Add(tc.until)),
							Offset: uint64(tc.offset),
						},
					},
					Query: &logs.LogQuery{
						QueryType: &logs.LogQuery_Admin{
							Admin: &logs.AdminLogQuery{
								AdminType: &logs.AdminLogQuery_Logql{
									Logql: `{app="testpach"}`,
								},
							},
						},
					},
				}
			)
			if err != nil {
				t.Fatalf("new test loki: %v", err)
			}
			t.Cleanup(func() {
				if err := aloki.Close(); err != nil {
					t.Fatalf("clean up loki: %v", err)
				}
			})
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			for i, offset := range tc.logs {
				var (
					timestamp = now.Add(offset)
					pachLog   = &pps.LogMessage{
						Ts:      timestamppb.New(timestamp),
						Message: fmt.Sprintf("log %d", i),
					}
					b   []byte
					log = &testloki.Log{
						Time: timestamp,
						Labels: map[string]string{
							"app": "testpach",
						},
					}
					err error
				)
				if b, err = protojson.Marshal(pachLog); err != nil {
					t.Fatal(err)
				}
				log.Message = string(b)
				if err = aloki.AddLog(ctx, log); err != nil {
					t.Fatal(err)
				}
			}
			if tc.wantHint != nil {
				req.WantPagingHint = true
			}
			require.NoError(t, ls.GetLogs(ctx, req, publisher), "GetLogs should succeed")
			var hints []*logs.PagingHint
			for _, resp := range publisher.responses {
				if hint := resp.GetPagingHint(); hint != nil {
					hints = append(hints, hint)
				}
			}
			if len(publisher.responses)-len(hints) != len(tc.want) {
				t.Fatalf("got %d responses; want %d", len(publisher.responses)-len(hints), len(tc.want))
			}
			for i, hint := range hints {
				if hint == nil || (hint.Older == nil && hint.Newer == nil) {
					t.Errorf("hint %d is empty", i)
				}
				if hint.Older != nil {
					if got, want := hint.Older.Filter.TimeRange.From.AsTime(), now.Add(tc.wantHint.olderFrom); !got.Equal(want) {
						t.Errorf("wanted older hint from = %v; got %v", want, got)
					}
					if got, want := hint.Older.Filter.TimeRange.Until.AsTime(), now.Add(tc.wantHint.olderUntil); !got.Equal(want) {
						t.Errorf("wanted older hint until = %v; got %v", want, got)
					}
					if got, want := hint.Older.Filter.TimeRange.Offset, uint64(tc.wantHint.olderOffset); got != want {
						t.Errorf("wanted older hint offset %d; got %d", want, got)
					}
				}
				if hint.Newer != nil {
					if got, want := hint.Newer.Filter.TimeRange.From.AsTime(), now.Add(tc.wantHint.newerFrom); !got.Equal(want) {
						t.Errorf("wanted newer hint from = %v; got %v", want, got)
					}
					if got, want := hint.Newer.Filter.TimeRange.Until.AsTime(), now.Add(tc.wantHint.newerUntil); !got.Equal(want) {
						t.Errorf("wanted newer hint until = %v; got %v", want, got)
					}
					if got, want := hint.Newer.Filter.TimeRange.Offset, uint64(tc.wantHint.newerOffset); got != want {
						t.Errorf("wanted newer hint offset %d; got %d", want, got)
					}
				}
			}
			for i := range tc.want {
				if want, got := now.Add(tc.want[i]), publisher.responses[i].GetLog().GetVerbatim().GetTimestamp().AsTime(); !want.Equal(got) {
					t.Errorf("expected item %d to be %v; got %v", i, want, got)
				}
			}

		})
	}
}

func TestWithRealLogs(t *testing.T) {
	testData := []struct {
		name  string
		query *logs.GetLogsRequest
		want  []*logs.GetLogsResponse
	}{
		{
			name: "edges logs",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_Admin{
						Admin: &logs.AdminLogQuery{
							AdminType: &logs.AdminLogQuery_Logql{
								Logql: `{suite="pachyderm",app="etcd"}`,
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					Limit: 1,
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(
							time.Date(2024, 04, 25, 17, 24, 0, 0, time.UTC),
						),
					},
				},
			},
			want: []*logs.GetLogsResponse{{
				ResponseType: &logs.GetLogsResponse_Log{
					Log: &logs.LogMessage{
						NativeTimestamp: timestamppb.New(time.Date(2024, 04, 25, 17, 24, 14, 67172000, time.UTC)),
						Verbatim: &logs.VerbatimLogMessage{
							Line:      []byte(`{"level":"info","ts":"2024-04-25T17:24:14.067172Z","caller":"flags/flag.go:113","msg":"recognized and used environment variable","variable-name":"ETCD_NAME","variable-value":"etcd-0"}`),
							Timestamp: timestamppb.New(time.Date(2024, 04, 25, 17, 24, 14, 67172000, time.UTC)),
						},
						PpsLogMessage: &pps.LogMessage{
							Ts: timestamppb.New(time.Date(2024, 04, 25, 17, 24, 14, 67172000, time.UTC)),
						},
						Object: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"level": {
									Kind: &structpb.Value_StringValue{
										StringValue: "info",
									},
								},
								"ts": {
									Kind: &structpb.Value_StringValue{
										StringValue: "2024-04-25T17:24:14.067172Z",
									},
								},
								"caller": {
									Kind: &structpb.Value_StringValue{
										StringValue: "flags/flag.go:113",
									},
								},
								"msg": {
									Kind: &structpb.Value_StringValue{
										StringValue: "recognized and used environment variable",
									},
								},
								"variable-name": {
									Kind: &structpb.Value_StringValue{
										StringValue: "ETCD_NAME",
									},
								},
								"variable-value": {
									Kind: &structpb.Value_StringValue{
										StringValue: "etcd-0",
									},
								},
							},
						},
					},
				},
			}},
		},
	}
	ctx := pctx.TestContext(t)
	l, err := testloki.New(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("testloki.New: %v", err)
	}
	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Fatalf("testloki.Close: %v", err)
		}
	})
	if err := filepath.Walk("testdata", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "called with error")
		}
		if !strings.HasSuffix(path, ".txt") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "open")
		}
		defer f.Close()
		if err := testloki.AddLogFile(ctx, f, l); err != nil {
			return errors.Wrap(err, "AddLogFile")
		}
		return nil
	}); err != nil {
		t.Fatalf("load logs: %v", err)
	}
	ls := logservice.LogService{
		GetLokiClient: func() (*loki.Client, error) {
			return l.Client, nil
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			p := new(testPublisher)
			if err := ls.GetLogs(ctx, test.query, p); err != nil {
				t.Fatalf("GetLogs: %v", err)
			}
			require.NoDiff(t, test.want, p.responses, []cmp.Option{protocmp.Transform()})
		})
	}
}
