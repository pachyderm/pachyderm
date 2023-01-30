package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
)

type lokiResult struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"`
}

type data struct {
	ResultType string       `json:"resultType"`
	Result     []lokiResult `json:"result"`
}
type response struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

func mustParseQuerystringInt64(r *http.Request, field string) int64 {
	x, err := strconv.ParseInt(r.URL.Query().Get(field), 10, 64)
	if err != nil {
		panic(err)
	}
	return x
}

type fakeLoki struct {
	entries     []loki.Entry // Must be sorted by time ascending.
	page        int          // Keep track of the current page.
	sleepAtPage int          // Which page to put the server to sleep. 0 means don't sleep.
}

func (l *fakeLoki) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Simulate the bug where Loki server hangs due to large logs.
	l.page++
	if l.sleepAtPage > 0 && l.page >= l.sleepAtPage {
		// wait for request to time out on purpose
		<-r.Context().Done()
		return
	}

	var (
		start, end time.Time
		limit      int
	)
	direction := r.URL.Query().Get("direction")
	if r.URL.Query().Get("start") != "" {
		start = time.Unix(0, mustParseQuerystringInt64(r, "start"))
	} else {
		start = time.Now().Add(-1 * time.Hour)
	}
	if r.URL.Query().Get("end") != "" {
		end = time.Unix(0, mustParseQuerystringInt64(r, "end"))
	} else {
		end = time.Now()
	}
	if r.URL.Query().Get("limit") != "" {
		limit = int(mustParseQuerystringInt64(r, "limit"))
	}

	if end.Before(start) {
		panic("end is before start")
	}
	if end.Sub(start) >= 721*time.Hour { // Not documented, but what a local Loki rejects.
		panic("query range too long")
	}

	var match []loki.Entry

	// From the logcli docs:
	// --from=FROM          Start looking for logs at this absolute time (inclusive)
	// --to=TO              Stop looking for logs at this absolute time (exclusive)
	// To is "end" and From is "start", so end is exclusive and start is inclusive.
	inRange := func(e loki.Entry) bool {
		return (e.Timestamp.After(start) || e.Timestamp.Equal(start)) && e.Timestamp.Before(end)
	}

	switch direction {
	case "FORWARD":
		for _, e := range l.entries {
			if inRange(e) {
				if len(match) >= limit {
					break
				}
				match = append(match, e)
			}
		}
	case "BACKWARD":
		for i := len(l.entries) - 1; i >= 0; i-- {
			e := l.entries[i]
			if inRange(e) {
				if len(match) >= limit {
					break
				}
				match = append(match, e)
			}
		}
	default:
		panic("invalid direction")
	}

	result := response{
		Status: "success",
		Data: data{
			ResultType: "streams",
			Result: []lokiResult{
				{
					Stream: map[string]string{"test": "stream"},
					Values: [][2]string{},
				},
			},
		},
	}
	for _, e := range match {
		result.Data.Result[0].Values = append(result.Data.Result[0].Values, [2]string{
			strconv.FormatInt(e.Timestamp.UnixNano(), 10),
			e.Line,
		})
	}

	content, err := json.Marshal(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(content); err != nil {
		panic(err)
	}
}

func TestQueryLoki(t *testing.T) {
	testData := []struct {
		name         string
		sleepAtPage  int
		buildEntries func() []loki.Entry
		buildWant    func() []int
	}{
		{
			name:         "no logs to return",
			buildEntries: func() []loki.Entry { return nil },
			buildWant:    func() []int { return nil },
		},
		{
			name: "all logs",
			buildEntries: func() []loki.Entry {
				var entries []loki.Entry
				for i := -99; i <= 0; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: time.Now().Add(time.Duration(-1) * time.Second),
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
		{
			name: "too many logs",
			buildEntries: func() []loki.Entry {
				start := time.Now()
				var entries []loki.Entry
				for i := -49999; i <= 0; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: start.Add(time.Duration(i) * time.Second),
						Line:      fmt.Sprintf("%v", i),
					})
				}
				return entries
			},
			buildWant: func() []int {
				var want []int
				for i := -29999; i <= 0; i++ {
					want = append(want, i)
				}
				return want
			},
		},
		{
			name: "big chunk of unchanging timestamps",
			buildEntries: func() []loki.Entry {
				start := time.Now()
				var entries []loki.Entry
				entries = append(entries, loki.Entry{
					Timestamp: start.Add(-2 * time.Second),
					Line:      "-2",
				})
				for i := 0; i < 40000; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: start.Add(-1 * time.Second),
						Line:      "-1",
					})
				}
				entries = append(entries, loki.Entry{
					Timestamp: start,
					Line:      "0",
				})
				return entries
			},
			buildWant: func() []int {
				var want []int
				want = append(want, -2)
				for i := 0; i < serverMaxLogs-1; i++ {
					want = append(want, -1)
				}
				want = append(want, 0)
				return want
			},
		},
		{
			name: "big chunk of unchanging timestamps at chunk boundary",
			buildEntries: func() []loki.Entry {
				start := time.Now()
				var entries []loki.Entry
				entries = append(entries, loki.Entry{
					Timestamp: start.Add(-2 * time.Second),
					Line:      "-2",
				})
				for i := 0; i < 40000; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: start.Add(-1 * time.Second),
						Line:      "-1",
					})
				}
				for i := 0; i < serverMaxLogs; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: start,
						Line:      "0",
					})
				}
				return entries
			},
			buildWant: func() []int {
				var want []int
				want = append(want, -2)
				for i := 0; i < serverMaxLogs; i++ {
					want = append(want, -1)
				}
				for i := 0; i < serverMaxLogs; i++ {
					want = append(want, 0)
				}
				return want
			},
		},
		{
			name:        "timeout on 1st page",
			sleepAtPage: 1,
			buildEntries: func() []loki.Entry {
				var entries []loki.Entry
				for i := -10000; i <= 0; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: time.Now().Add(time.Duration(-1) * time.Second),
						Line:      fmt.Sprintf("%v", i),
					})
				}
				return entries
			},
			buildWant: func() []int {
				// server should've timed out right away, so no results
				return nil
			},
		},
		{
			name:        "timeout on 2nd page",
			sleepAtPage: 2,
			buildEntries: func() []loki.Entry {
				var entries []loki.Entry
				for i := -10000; i <= 0; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: time.Now().Add(time.Duration(-1) * time.Second),
						Line:      fmt.Sprintf("%v", i),
					})
				}
				return entries
			},
			buildWant: func() []int {
				// expect only the first page due to server timing out
				var want []int
				for i := -999; i <= 0; i++ {
					want = append(want, i)
				}
				return want
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			entries := test.buildEntries()
			want := test.buildWant()

			s := httptest.NewServer(&fakeLoki{
				entries:     entries,
				sleepAtPage: test.sleepAtPage,
			})
			d := &debugServer{
				env: &serviceenv.TestServiceEnv{
					LokiClient: &loki.Client{Address: s.URL},
				},
			}

			var got []int
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			out, err := d.queryLoki(ctx, `{foo="bar"}`)
			if err != nil {
				t.Fatalf("query loki: %v", err)
			}
			for i, l := range out {
				x, err := strconv.ParseInt(l.Entry.Line, 10, 64)
				if err != nil {
					t.Errorf("parse log line %d: %v", i, err)
				}
				got = append(got, int(x))
			}

			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf(`result differs:
          first |   last |    len
       +--------+--------+-------
   got | %6d | %6d | %6d
  want | %6d | %6d | %6d

slice samples:
   got: %v ... %v
  want: %v ... %v`,
					got[0], got[len(got)-1], len(got),
					want[0], want[len(want)-1], len(want),
					got[:4], got[len(got)-4:],
					want[:4], want[len(want)-4:])

				if testing.Verbose() {
					t.Logf("returned lines:\n%v", diff)
				}
			}
		})
	}
}

func TestQuoteLogQL(t *testing.T) {
	for s, q := range map[string]string{
		"abc":  `"abc"`,
		"a'bc": `"a'bc"`,
		`a"bc`: `"a\"bc"`,
		`a
bc`: `"a\nbc"`,
		`ßþ…`: `"ßþ…"`,
	} {
		if quoteLogQLStreamSelector(s) != q {
			t.Errorf("expected quoteLogQL(%q) = %q; got %q", s, q, quoteLogQLStreamSelector(s))
		}
	}
}
