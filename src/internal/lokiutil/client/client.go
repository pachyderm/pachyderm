package client

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
)

// ** Why this is here **
// We use a stripped down version of the loki client as importing
// the main client locks us to old module deps like k8s.io/client-go
var (
	queryRangePath = "/loki/api/v1/query_range"
)

// Client holds configuration for the loki
type Client struct {
	Address string
}

var lokiClient = &http.Client{
	Transport: promutil.InstrumentRoundTripper("loki", http.DefaultTransport),
}

// QueryRange queries Loki in a given time range.
func (c *Client) QueryRange(ctx context.Context, queryStr string, limit int, start, end time.Time, direction string, step, interval time.Duration, quiet bool) (*QueryResponse, error) {
	params := newQueryStringBuilder()
	params.SetString("query", queryStr)
	if limit > 0 {
		params.SetInt32("limit", limit)
	}
	if !start.IsZero() {
		params.SetInt("start", start.UnixNano())
	}
	if !end.IsZero() {
		params.SetInt("end", end.UnixNano())
	}
	params.SetString("direction", direction)

	// The step is optional, so we do set it only if provided,
	// otherwise we do leverage on the API defaults
	if step != 0 {
		params.SetFloat("step", step.Seconds())
	}

	if interval != 0 {
		params.SetFloat("interval", interval.Seconds())
	}

	return c.doQuery(ctx, queryRangePath, params.Encode(), quiet)
}

func (c *Client) doQuery(ctx context.Context, path string, query string, quiet bool) (*QueryResponse, error) {
	var err error
	var r QueryResponse

	if err = c.doRequest(ctx, path, query, quiet, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (c *Client) doRequest(ctx context.Context, path, query string, quiet bool, out interface{}) error {
	us, err := buildURL(c.Address, path, query)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", us, nil)
	if err != nil {
		return errors.EnsureStack(err)
	}

	resp, err := lokiClient.Do(req)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.EnsureStack(errors.Errorf("error response from loki: %v (body: %q); additionally, reading body: %v", resp.Status, body, err))
		}
		return errors.EnsureStack(errors.Errorf("error response from loki: %v (body: %q)", resp.Status, body))
	}
	return errors.EnsureStack(json.NewDecoder(resp.Body).Decode(out))
}

func buildURL(u, p, q string) (string, error) {
	url, err := url.Parse(u)
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	url.Path = path.Join(url.Path, p)
	url.RawQuery = q
	return url.String(), nil
}

type lokiConfig struct {
	LimitsConfig struct {
		MaxQueryLength string `yaml:"max_query_length"`
	} `yaml:"limits_config"`
}

// MaxQueryLength returns the server’s configured max query length.  It returns
// 0 if there is none.
func (c *Client) MaxQueryLength(ctx context.Context) (time.Duration, error) {
	u, err := buildURL(c.Address, "/config", "")
	if err != nil {
		return 0, errors.Wrap(err, "could not build config URL")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return 0, errors.Wrap(err, "could not create HTTP request")
	}
	resp, err := lokiClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "could not make HTTP request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return 0, errors.Wrapf(err, "do not know how to handle HTTP status %d (%s)", resp.StatusCode, resp.Status)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "could not read response body")
	}
	var config lokiConfig
	if err := yaml.Unmarshal(b, &config); err != nil {
		return 0, errors.Wrap(err, "could not parse YAML response")
	}
	if config.LimitsConfig.MaxQueryLength == "" {
		return 0, nil
	}
	mql, err := model.ParseDuration(config.LimitsConfig.MaxQueryLength)
	if err != nil {
		return 0, errors.Wrap(err, "could not parse server’s returned max query length")
	}
	return time.Duration(mql), nil
}
