package lokiutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
)

type FakeServer struct {
	Entries     []loki.Entry // Must be sorted by time ascending.
	Page        int          // Keep track of the current page.
	SleepAtPage int          // Which page to put the server to sleep. 0 means don't sleep.
}

type response struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

type data struct {
	ResultType string       `json:"resultType"`
	Result     []lokiResult `json:"result"`
}

type lokiResult struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"`
}

func mustParseQuerystringInt64(r *http.Request, field string) int64 {
	x, err := strconv.ParseInt(r.URL.Query().Get(field), 10, 64)
	if err != nil {
		panic(err)
	}
	return x
}

func (l *FakeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Simulate the bug where Loki server hangs due to large logs.
	l.Page++
	if l.SleepAtPage > 0 && l.Page >= l.SleepAtPage {
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
	} else {
		limit = 5000
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

	switch strings.ToLower(direction) {
	case "forward":
		for _, e := range l.Entries {
			if inRange(e) {
				if len(match) >= limit {
					break
				}
				match = append(match, e)
			}
		}
	case "backward":
		for i := len(l.Entries) - 1; i >= 0; i-- {
			e := l.Entries[i]
			if inRange(e) {
				if len(match) >= limit {
					break
				}
				match = append(match, e)
			}
		}
	default:
		http.Error(w, fmt.Sprintf("invalid direction %s", direction), http.StatusBadRequest)
		return
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
