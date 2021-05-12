package lokiutil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"time"
)

var (
	queryRangePath = "/loki/api/v1/query_range"
	lokiAddress    = "http://localhost:3100" //TODO Make configurable for on cluster
)

//	QueryRange(queryStr string, limit int, from, through time.Time, direction logproto.Direction, step, interval time.Duration, quiet bool) (*loghttp.QueryResponse, error)
func QueryRangeInternal(ctx context.Context, queryStr string, limit int, start, end time.Time, direction string, step, interval time.Duration, quiet bool) (*QueryResponse, error) {
	params := NewQueryStringBuilder()
	params.SetString("query", queryStr)
	params.SetInt32("limit", limit)
	params.SetInt("start", start.UnixNano())
	params.SetInt("end", end.UnixNano())
	params.SetString("direction", direction)

	// The step is optional, so we do set it only if provided,
	// otherwise we do leverage on the API defaults
	if step != 0 {
		params.SetFloat("step", step.Seconds())
	}

	if interval != 0 {
		params.SetFloat("interval", interval.Seconds())
	}

	return doQuery(ctx, queryRangePath, params.Encode(), quiet)
}

func doQuery(ctx context.Context, path string, query string, quiet bool) (*QueryResponse, error) {
	var err error
	var r QueryResponse

	if err = doRequest(ctx, path, query, quiet, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func doRequest(ctx context.Context, path, query string, quiet bool, out interface{}) error {
	us, err := buildURL(lokiAddress, path, query) //TODO pass through address
	if err != nil {
		return err
	}
	//TODO pass through context with http.NewRequest("GET", "http://www.yahoo.co.jp", nil)
	// https://gist.github.com/superbrothers/dae0030c151d1f3c24311df77405169b
	req, err := http.Get(us)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	return json.NewDecoder(req.Body).Decode(out)
}

func buildURL(u, p, q string) (string, error) {
	url, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	url.Path = path.Join(url.Path, p)
	url.RawQuery = q
	return url.String(), nil
}
