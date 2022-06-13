package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
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
	entries []loki.Entry // Must be sorted by time ascending.
}

func (l *fakeLoki) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
				for i := -99; i >= 0; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: time.Now().Add(time.Duration(-1) * time.Second),
						Line:      fmt.Sprintf("%v", i),
					})
				}
				return entries
			},
			buildWant: func() []int {
				var want []int
				for i := -99; i >= 0; i++ {
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
				for i := 0; i < 4999; i++ {
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
				for i := 0; i < 5000; i++ {
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
				for i := 0; i < 5000; i++ {
					want = append(want, -1)
				}
				for i := 0; i < 5000; i++ {
					want = append(want, 0)
				}
				return want
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			entries := test.buildEntries()
			want := test.buildWant()

			s := httptest.NewServer(&fakeLoki{entries: entries})
			d := &debugServer{
				env: &serviceenv.TestServiceEnv{
					LokiClient: &loki.Client{Address: s.URL},
				},
			}

			var got []int
			out, err := d.queryLoki(context.Background(), `{foo="bar"}`)
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

func TestPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, _ := minikubetestenv.AcquireCluster(t)

	numCommits := 5

	// Load some very basic dummy data into the database.
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl create repo {{.repo}}

		# Create a commit and put some data in it
		for i in $(seq 1 {{.commits}}); do
			commit=$(pachctl start commit {{.repo}}@master)
			echo "file contents" | pachctl put file {{.repo}}@${commit}:/file -f -
			pachctl finish commit {{.repo}}@${commit}
		done;
	`, "repo", tu.UniqueString("TestCommit-repo"), "commits", strconv.Itoa(numCommits),
	).Run())

	require.NoError(t, tu.PachctlBashCmd(t, c, `
		returnCode=0

		pachctl debug dump /tmp/{{.dumpfile}}
		cd /tmp
		tar -xf {{.dumpfile}} postgres

		# ls should return here with a return code of 0 if all the files are found 
		cd postgres
		ls activities.txt  row-counts.txt table-sizes.txt

		# compare commits in database to what we committed
		rowCountCommits=$(cat row-counts.txt | grep commits | tr -s " " | cut -d " " -f 3)
		if [[ "${rowCountCommits}" -ne "{{.commits}}" ]]; then
			returnCode=1 #return at the end to ensure we can cleanup temp files
		fi	

		cd ..
		rm -rf {{.dumpfile}} postgres

		exit ${returnCode}
	`, "dumpfile", tu.UniqueString("out.tar.gz"), "commits", strconv.Itoa(numCommits),
	).Run())
}
