package logs_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"io/fs"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/itchyny/gojq"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	logservice "github.com/pachyderm/pachyderm/v2/src/server/logs"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/testloki"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
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
		projectName     = uuid.UniqueString("pipelineProject")
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
		for k, v := range labels {
			object.Fields["#"+k] = structpb.NewStringValue(v)
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
	olderUntil  time.Duration
	olderOffset uint
	newerFrom   time.Duration
	newerOffset uint
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
		testLogs = []time.Duration{-700*time.Hour - time.Second, -1 * time.Hour, -1 * time.Minute, -1 * time.Second, time.Minute, time.Hour}
		want     = []time.Duration{-700*time.Hour - time.Second, -1 * time.Hour, -1 * time.Minute, -1 * time.Second}
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
				Message: fmt.Sprintf("log %d (%v)", i, offset.String()),
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
			t.Errorf("hint %d is empty (%v)", i, hint.String())
		}
		// Since we cannot know the server-side time, and the default is
		// set server side, check that there is a consistent delta.
		//
		// NOTE: this assumes that a single hint contains both older and
		// newer hints; it won’t handle things properly if they are
		// split up.  Doing that is trickier than it is probably worth
		// right now.
		if hint.Older != nil {
			if got, want := hint.Older.Filter.TimeRange.Until.AsTime(), time.Date(2019, 12, 31, 23, 59, 59, 999999999, time.UTC); got != want {
				t.Errorf("older hint:\n  got: %v\n want: %v", got.Format(time.RFC3339Nano), want.Format(time.RFC3339Nano))
			}
		}
		if hint.Newer != nil {
			got := hint.Newer.Filter.TimeRange.From.AsTime()
			want := time.Now()
			if want.Sub(got) > time.Minute {
				t.Errorf("newer paging hint not close enough to present time: %v", got.Format(time.RFC3339Nano))
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
		"from is inclusive; until is inclusive": {
			logs:  []time.Duration{time.Nanosecond, time.Second, time.Second * 3 / 2, time.Second * 2, time.Second * 12, time.Second * 14},
			limit: 0,
			from:  time.Second,
			until: time.Second*12 - time.Nanosecond,
			want:  []time.Duration{time.Second, time.Second * 3 / 2, time.Second * 2},
		},
		"from after until is backwards": {
			logs:  []time.Duration{time.Nanosecond, time.Second, time.Second * 3 / 2, time.Second * 2, time.Second * 12, time.Second * 14},
			limit: 0,
			from:  time.Second * 11,
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
				olderUntil: time.Second * -10,
				newerFrom:  time.Second * 12,
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
				olderUntil:  time.Second * -10,
				newerFrom:   time.Second*2 + time.Nanosecond,
				newerOffset: 0,
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
			logs:  []time.Duration{1, time.Second * 2, time.Second * 3},
			limit: 1,
			from:  time.Second*2 + time.Nanosecond,
			until: time.Second * 13,
			want:  []time.Duration{time.Second * 3},
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
				olderUntil: time.Second * -10,
				newerFrom:  time.Second * 12,
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
				olderUntil: time.Second * -10,
				newerFrom:  time.Second * 3,
			},
		},
		"offset works": {
			logs:  []time.Duration{time.Second * 2, time.Second * 3, time.Second * 3},
			limit: 2,
			from:  time.Second,
			until: time.Second * 12,
			want:  []time.Duration{time.Second * 2, time.Second * 3},
			wantHint: &hintCase{
				olderUntil:  time.Second * -10,
				newerFrom:   time.Second * 3,
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
				olderUntil: time.Second * -2,
				newerFrom:  time.Second * 10,
			},
		},
		"backwards older hint works": {
			logs:  []time.Duration{time.Second * -3, time.Second * -2, time.Second * -1, 0, time.Second, time.Second * 2},
			limit: 0,
			from:  time.Second * -2,
			until: time.Second * -3,
			want:  []time.Duration{time.Second * -2, time.Second * -3},
			wantHint: &hintCase{
				olderUntil: time.Second * -3,
				newerFrom:  time.Second * -1,
			},
		},
		"backwards older hint with a limit is correct": {
			logs:  []time.Duration{time.Second * -3, time.Second * -2, time.Second * -1, 0, time.Second, time.Second * 2},
			limit: 2,
			from:  time.Second * 10,
			until: time.Second * -1,
			want:  []time.Duration{time.Second * 2, time.Second},
			wantHint: &hintCase{
				olderUntil: time.Second * -10,
				newerFrom:  time.Second * 10,
			},
		},
		"backwards older hint with a limit works": {
			logs:  []time.Duration{time.Second * -3, time.Second * -2, time.Second * -1, 0, time.Second, time.Second * 2},
			limit: 2,
			from:  time.Second,
			until: time.Second * -10,
			want:  []time.Duration{time.Second, 0},
			wantHint: &hintCase{
				olderUntil: time.Second * -10,
				newerFrom:  time.Second,
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
			if tc.from != 0 {
				req.Filter.TimeRange.From = timestamppb.New(now.Add(tc.from))
			}
			if tc.until != 0 {
				req.Filter.TimeRange.Until = timestamppb.New(now.Add(tc.until))
			}

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
					if got, want := hint.Older.Filter.TimeRange.Until.AsTime(), now.Add(tc.wantHint.olderUntil); math.Abs(float64(got.Sub(want))) > float64(time.Minute) {
						t.Errorf("wanted older hint until = %v; got %v", want, got)
					}
					if got, want := hint.Older.Filter.TimeRange.Offset, uint64(tc.wantHint.olderOffset); got != want {
						t.Errorf("wanted older hint offset %d; got %d", want, got)
					}
				}
				if hint.Newer != nil {
					if got, want := hint.Newer.Filter.TimeRange.From.AsTime(), now.Add(tc.wantHint.newerFrom); math.Abs(float64(got.Sub(want))) > float64(time.Minute) {
						t.Errorf("wanted newer hint from = %v; got %v", want, got)
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

func transformProto[T proto.Message](name string, filter func(T) T) cmp.Option {
	return protocmp.FilterMessage(*new(T), cmpopts.AcyclicTransformer(name, func(in any) any {
		x := in.(protocmp.Message).Unwrap().(T)
		return filter(x)
	}))
}

// onlyCompareObject ignores all fields on logs.LogMessage except "object".
var onlyCompareObject = transformProto("onlyCompareObject", func(lm *logs.LogMessage) *logs.LogMessage {
	return &logs.LogMessage{
		Object: lm.GetObject(),
	}
})

// onlyTimeRangeInPagingHint ignores all fields on a paging hint that aren't related to time ranges.
var onlyTimeRangeInPagingHint = transformProto("onlyTimeRangeInPagingHint", func(ph *logs.PagingHint) *logs.PagingHint {
	return &logs.PagingHint{
		Older: &logs.GetLogsRequest{
			Filter: &logs.LogFilter{
				TimeRange: ph.GetOlder().GetFilter().GetTimeRange(),
			},
		},
		Newer: &logs.GetLogsRequest{
			Filter: &logs.LogFilter{
				TimeRange: ph.GetNewer().GetFilter().GetTimeRange(),
			},
		},
	}
})

var sortByNativeTimestamp = cmpopts.SortSlices(func(x, y *logs.GetLogsResponse) bool {
	return x.GetLog().GetNativeTimestamp().AsTime().Before(y.GetLog().GetNativeTimestamp().AsTime())
})

// jqObject runs the provided JQ program on all structpb.Struct objects (useful for transforming
// GetLogsResponse.Log.Object).
func jqObject(program string) cmp.Option {
	q, err := gojq.Parse(program)
	if err != nil {
		panic(fmt.Sprintf("parse jq program %q: %v", program, err))
	}
	jq, err := gojq.Compile(q)
	if err != nil {
		panic(fmt.Sprintf("compile jq program %q: %v", program, err))
	}
	return transformProto("jqObject", func(s *structpb.Struct) *structpb.Struct {
		iter := jq.Run(s.AsMap())
		if result, ok := iter.Next(); ok {
			switch x := result.(type) {
			case map[string]any:
				rs, err := structpb.NewStruct(x)
				if err != nil {
					panic(fmt.Sprintf("NewStruct(%v): %v", x, err))
				}
				return rs
			default:
				panic(fmt.Sprintf("jq program produced invalid output type: %#v; want map[string]any", result))
			}
		}
		panic("jq program failed to produce output")

	})
}

// prettyTimes translates timestamps to RFC3339Nano strings.
var prettyTimes = protocmp.FilterMessage(new(timestamppb.Timestamp), cmpopts.AcyclicTransformer("prettyTimes", func(in any) any {
	x := in.(protocmp.Message).Unwrap().(*timestamppb.Timestamp)
	if x == nil {
		return "<nil>"
	}
	return x.AsTime().Format(time.RFC3339Nano)
}))

// jsonLog creates a new JSON log message suitable for comparison with a real log via
// onlyCompareObject.
func jsonLog(x map[string]any) *logs.GetLogsResponse {
	s, err := structpb.NewStruct(x)
	if err != nil {
		panic(fmt.Sprintf("unmarshal %v json into google.protobuf.Struct: %v", x, err))
	}
	return &logs.GetLogsResponse{
		ResponseType: &logs.GetLogsResponse_Log{
			Log: &logs.LogMessage{
				Object: s,
			},
		},
	}
}

func TestWithRealLogs(t *testing.T) {
	testData := []struct {
		name  string
		query *logs.GetLogsRequest
		want  []*logs.GetLogsResponse
		opts  []cmp.Option
	}{
		{
			name: "etcd logs",
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
				WantPagingHint: true,
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
								"#app": {
									Kind: &structpb.Value_StringValue{
										StringValue: "etcd",
									},
								},
								"#apps_kubernetes_io_pod_index": {
									Kind: &structpb.Value_StringValue{
										StringValue: "0",
									},
								},
								"#container": {
									Kind: &structpb.Value_StringValue{
										StringValue: "etcd",
									},
								},
								"#controller_revision_hash": {
									Kind: &structpb.Value_StringValue{
										StringValue: "etcd-5c748dc64f",
									},
								},
								"#filename": {
									Kind: &structpb.Value_StringValue{
										StringValue: "/var/log/pods/default_etcd-0_d80ed43c-7439-4a97-a905-7e035e34d683/etcd/0.log",
									},
								},
								"#job": {
									Kind: &structpb.Value_StringValue{
										StringValue: "default/etcd",
									},
								},
								"#namespace": {
									Kind: &structpb.Value_StringValue{
										StringValue: "default",
									},
								},
								"#node_name": {
									Kind: &structpb.Value_StringValue{
										StringValue: "pach-control-plane",
									},
								},
								"#pod": {
									Kind: &structpb.Value_StringValue{
										StringValue: "etcd-0",
									},
								},
								"#statefulset_kubernetes_io_pod_name": {
									Kind: &structpb.Value_StringValue{
										StringValue: "etcd-0",
									},
								},
								"#stream": {
									Kind: &structpb.Value_StringValue{
										StringValue: "stderr",
									},
								},
								"#suite": {
									Kind: &structpb.Value_StringValue{
										StringValue: "pachyderm",
									},
								},
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
			},
				{
					ResponseType: &logs.GetLogsResponse_PagingHint{
						PagingHint: &logs.PagingHint{
							Older: &logs.GetLogsRequest{
								Filter: &logs.LogFilter{
									TimeRange: &logs.TimeRangeLogFilter{
										Until: timestamppb.New(time.Date(2024, 04, 25, 17, 23, 59, 999_999_999, time.UTC)),
									},
									Limit: 1,
								},
								Query: &logs.LogQuery{
									QueryType: &logs.LogQuery_Admin{
										Admin: &logs.AdminLogQuery{
											AdminType: &logs.AdminLogQuery_Logql{
												Logql: `{suite="pachyderm",app="etcd"}`,
											},
										},
									},
								},
								WantPagingHint: true,
							},
							Newer: &logs.GetLogsRequest{
								Filter: &logs.LogFilter{
									TimeRange: &logs.TimeRangeLogFilter{
										From: timestamppb.New(time.Date(2024, 04, 25, 17, 24, 14, 67172001, time.UTC)),
									},
									Limit: 1,
								},
								Query: &logs.LogQuery{
									QueryType: &logs.LogQuery_Admin{
										Admin: &logs.AdminLogQuery{
											AdminType: &logs.AdminLogQuery_Logql{
												Logql: `{suite="pachyderm",app="etcd"}`,
											},
										},
									},
								},
								WantPagingHint: true,
							},
						},
					},
				},
			},
		},
		{
			name: "simple logs mid-stream",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_Admin{
						Admin: &logs.AdminLogQuery{
							AdminType: &logs.AdminLogQuery_App{
								App: "simple",
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 970, time.UTC)),
					},
					Limit: 10,
				},
				WantPagingHint: true,
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}"), onlyTimeRangeInPagingHint},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{"message": "970", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "971", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "972", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "973", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "974", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "975", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "976", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "977", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "978", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "979", "#suite": "pachyderm", "#app": "simple"}),
				{
					ResponseType: &logs.GetLogsResponse_PagingHint{
						PagingHint: &logs.PagingHint{
							Older: &logs.GetLogsRequest{
								Filter: &logs.LogFilter{
									TimeRange: &logs.TimeRangeLogFilter{
										Until: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 969, time.UTC)),
									},
								},
							},
							Newer: &logs.GetLogsRequest{
								Filter: &logs.LogFilter{
									TimeRange: &logs.TimeRangeLogFilter{
										From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 980, time.UTC)),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "simple logs mid-stream, errors only",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_Admin{
						Admin: &logs.AdminLogQuery{
							AdminType: &logs.AdminLogQuery_App{
								App: "simple",
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 970, time.UTC)),
					},
					Limit: 3,
					Level: logs.LogLevel_LOG_LEVEL_ERROR,
				},
				WantPagingHint: true,
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}"), onlyTimeRangeInPagingHint},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{"message": "971", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "974", "#suite": "pachyderm", "#app": "simple"}),
				jsonLog(map[string]any{"message": "977", "#suite": "pachyderm", "#app": "simple"}),
				{
					ResponseType: &logs.GetLogsResponse_PagingHint{
						PagingHint: &logs.PagingHint{
							Older: &logs.GetLogsRequest{
								Filter: &logs.LogFilter{
									TimeRange: &logs.TimeRangeLogFilter{
										Until: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 969, time.UTC)),
									},
									Level: logs.LogLevel_LOG_LEVEL_ERROR,
								},
							},
							Newer: &logs.GetLogsRequest{
								Filter: &logs.LogFilter{
									TimeRange: &logs.TimeRangeLogFilter{
										From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 978, time.UTC)),
									},
									Level: logs.LogLevel_LOG_LEVEL_ERROR,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "pipeline logs",
			query: &logs.GetLogsRequest{
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
				Filter: &logs.LogFilter{
					Limit: 2,
					// Implicitly filtered to INFO and above.
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "connecting to postgres",
				}),
				jsonLog(map[string]any{
					"message": "started transform spawner process",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "pipeline logs at debug level",
			query: &logs.GetLogsRequest{
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
				Filter: &logs.LogFilter{
					Limit: 2,
					Level: logs.LogLevel_LOG_LEVEL_DEBUG,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "version info",
				}),
				jsonLog(map[string]any{
					"message": "serviceenv: span start",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "user code logs from a pipeline",
			query: &logs.GetLogsRequest{
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
				Filter: &logs.LogFilter{
					UserLogsOnly: true,
					Limit:        2,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.",
				}),
				jsonLog(map[string]any{
					"message": "  warnings.warn('Matplotlib is building the font cache using fc-list. This may take a moment.')",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "job + datum",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_JobDatum{
								JobDatum: &logs.JobDatumLogQuery{
									Job:   "14a8aecd0b944adb96e9ab1ad06cf29e",
									Datum: "3c726887e69fb82e628366072e57087868f7f42b29613ac50ffaa692e5a78d5c",
								},
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					// TODO(jrockway): |= in Loki is INCREDIBLY slow for some reason.
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 6, 13, 3, 0, 40, 0, time.UTC)),
					},
					Limit: 2,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "beginning to run user code",
				}),
				jsonLog(map[string]any{
					"message": "/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "user code logs from a job/datum",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_JobDatum{
								JobDatum: &logs.JobDatumLogQuery{
									Job:   "14a8aecd0b944adb96e9ab1ad06cf29e",
									Datum: "3c726887e69fb82e628366072e57087868f7f42b29613ac50ffaa692e5a78d5c",
								},
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 6, 13, 3, 0, 40, 0, time.UTC)),
					},
					UserLogsOnly: true,
					Limit:        2,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.",
				}),
				jsonLog(map[string]any{
					"message": "  warnings.warn('Matplotlib is building the font cache using fc-list. This may take a moment.')",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "datum",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_Datum{
								Datum: "3c726887e69fb82e628366072e57087868f7f42b29613ac50ffaa692e5a78d5c",
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 6, 13, 3, 0, 40, 0, time.UTC)),
					},
					Limit: 2,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "beginning to run user code",
				}),
				jsonLog(map[string]any{
					"message": "/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "user code logs from a datum",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_Datum{
								Datum: "3c726887e69fb82e628366072e57087868f7f42b29613ac50ffaa692e5a78d5c",
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 6, 13, 3, 0, 40, 0, time.UTC)),
					},
					UserLogsOnly: true,
					Limit:        2,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.",
				}),
				jsonLog(map[string]any{
					"message": "  warnings.warn('Matplotlib is building the font cache using fc-list. This may take a moment.')",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "project",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_Project{
								Project: "default",
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)),
					},
					Limit: 2,
					Level: logs.LogLevel_LOG_LEVEL_DEBUG,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "version info",
				}),
				jsonLog(map[string]any{
					"message": "serviceenv: span start",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "user code logs from a project",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_Project{
								Project: "default",
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)),
					},
					UserLogsOnly: true,
					Limit:        2,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.",
				}),
				jsonLog(map[string]any{
					"message": "  warnings.warn('Matplotlib is building the font cache using fc-list. This may take a moment.')",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "job",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_Job{
								Job: "14a8aecd0b944adb96e9ab1ad06cf29e",
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)),
					},
					Limit: 2,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "updating job info, state: JOB_STARTING",
				}),
				jsonLog(map[string]any{
					"message": "started waiting for job inputs",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "user code logs from a pipeline job",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_PipelineJob{
								PipelineJob: &logs.PipelineJobLogQuery{
									Pipeline: &logs.PipelineLogQuery{
										Project:  "default",
										Pipeline: "edges",
									},
									Job: "14a8aecd0b944adb96e9ab1ad06cf29e",
								},
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)),
					},
					Limit:        2,
					UserLogsOnly: true,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.",
				}),
				jsonLog(map[string]any{
					"message": "  warnings.warn('Matplotlib is building the font cache using fc-list. This may take a moment.')",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "job",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_PipelineJob{
								PipelineJob: &logs.PipelineJobLogQuery{
									Pipeline: &logs.PipelineLogQuery{
										Project:  "default",
										Pipeline: "edges",
									},
									Job: "14a8aecd0b944adb96e9ab1ad06cf29e",
								},
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)),
					},
					Limit: 2,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "updating job info, state: JOB_STARTING",
				}),
				jsonLog(map[string]any{
					"message": "started waiting for job inputs",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "user code logs from a pipeline job",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_Job{
								Job: "14a8aecd0b944adb96e9ab1ad06cf29e",
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)),
					},
					Limit:        2,
					UserLogsOnly: true,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.",
				}),
				jsonLog(map[string]any{
					"message": "  warnings.warn('Matplotlib is building the font cache using fc-list. This may take a moment.')",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "datum logs for a pipeline, with user filter",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_PipelineDatum{
								PipelineDatum: &logs.PipelineDatumLogQuery{
									Pipeline: &logs.PipelineLogQuery{
										Project:  "default",
										Pipeline: "edges",
									},
									Datum: "3c726887e69fb82e628366072e57087868f7f42b29613ac50ffaa692e5a78d5c",
								},
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)),
					},
					Limit:        2,
					UserLogsOnly: true,
				},
			},
			want: []*logs.GetLogsResponse{
				jsonLog(map[string]any{
					"message": "/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.",
				}),
				jsonLog(map[string]any{
					"message": "  warnings.warn('Matplotlib is building the font cache using fc-list. This may take a moment.')",
				}),
			},
			opts: []cmp.Option{onlyCompareObject, jqObject("{message}")},
		},
		{
			name: "datum logs for a pipeline that doesn't contain that datum",
			query: &logs.GetLogsRequest{
				Query: &logs.LogQuery{
					QueryType: &logs.LogQuery_User{
						User: &logs.UserLogQuery{
							UserType: &logs.UserLogQuery_PipelineDatum{
								PipelineDatum: &logs.PipelineDatumLogQuery{
									Pipeline: &logs.PipelineLogQuery{
										Project:  "default",
										Pipeline: "montage",
									},
									Datum: "3c726887e69fb82e628366072e57087868f7f42b29613ac50ffaa692e5a78d5c",
								},
							},
						},
					},
				},
				Filter: &logs.LogFilter{
					TimeRange: &logs.TimeRangeLogFilter{
						From: timestamppb.New(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)),
					},
					Limit: 2,
				},
			},
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
		defer f.Close() //nolint:errcheck
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
			opts := []cmp.Option{protocmp.Transform(), prettyTimes}
			opts = append(opts, test.opts...)
			require.NoDiff(t, test.want, p.responses, opts)
		})
	}

	t.Run("FollowPagingHints_Forward", followPagingHintsTest(
		&ls,
		&logs.GetLogsRequest{
			Query: &logs.LogQuery{
				QueryType: &logs.LogQuery_Admin{
					Admin: &logs.AdminLogQuery{
						AdminType: &logs.AdminLogQuery_App{
							App: "simple",
						},
					},
				},
			},
			Filter: &logs.LogFilter{
				TimeRange: &logs.TimeRangeLogFilter{
					From: timestamppb.New(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
				Limit: 10,
			},
			WantPagingHint: true,
		},
		func(x *logs.PagingHint) *logs.GetLogsRequest { return x.Newer }),
	)
	t.Run("FollowPagingHints_Backwards", followPagingHintsTest(
		&ls,
		&logs.GetLogsRequest{
			Query: &logs.LogQuery{
				QueryType: &logs.LogQuery_Admin{
					Admin: &logs.AdminLogQuery{
						AdminType: &logs.AdminLogQuery_App{
							App: "simple",
						},
					},
				},
			},
			Filter: &logs.LogFilter{
				TimeRange: &logs.TimeRangeLogFilter{
					Until: timestamppb.New(time.Date(2024, 5, 2, 0, 0, 0, 0, time.UTC)),
				},
				Limit: 10,
			},
			WantPagingHint: true,
		},
		func(x *logs.PagingHint) *logs.GetLogsRequest { return x.Older }),
	)
}

func followPagingHintsTest(ls *logservice.LogService, initial *logs.GetLogsRequest, next func(*logs.PagingHint) *logs.GetLogsRequest) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		ctx := pctx.TestContext(t)
		q := initial
		var want []*logs.GetLogsResponse
		for i := 0; i < 1000; i++ {
			want = append(want, jsonLog(map[string]any{
				"#app":    "simple",
				"#suite":  "pachyderm",
				"time":    time.Date(2024, 5, 1, 0, 0, 0, i, time.UTC).Format(time.RFC3339Nano),
				"message": strconv.Itoa(i),
			}))
		}

		var got []*logs.GetLogsResponse
		var terminated bool
		for i := 0; i < 1000; i++ { // avoid infinite loop
			p := new(testPublisher)
			if err := ls.GetLogs(ctx, q, p); err != nil {
				t.Fatalf("GetLogs: %v", err)
			}
			if len(p.responses) == 0 {
				break
			}
			var gotLogs bool
			var gotPagingHint bool
			for _, r := range p.responses {
				switch x := r.ResponseType.(type) {
				case *logs.GetLogsResponse_PagingHint:
					if gotPagingHint {
						t.Errorf("got duplicate paging hint %v", x.PagingHint)
					}
					q = next(x.PagingHint)
					gotPagingHint = true
					log.Info(ctx, "next page", zap.Int("i", i), log.Proto("timeRange", q.GetFilter().GetTimeRange()))
				case *logs.GetLogsResponse_Log:
					gotLogs = true
					got = append(got, r)
				}
			}
			if !gotPagingHint {
				t.Errorf("did not get a paging hint on batch %v", i)
				terminated = true
				break
			}
			if !gotLogs {
				terminated = true
				break
			}
		}
		if !terminated {
			t.Errorf("logs loop should have terminated on its own, but didn't")
		}
		require.NoDiff(t, want, got, []cmp.Option{sortByNativeTimestamp, protocmp.Transform(), prettyTimes, jqObject("del(.severity)"), onlyCompareObject})
	}
}
