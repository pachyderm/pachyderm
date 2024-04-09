package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
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

type tailResponse struct {
	Streams []tailStream `json:"streams"`
}

type tailStream struct {
	Stream map[string]string `json:"stream"`
	Values []tailValue       `json:"values"`
}

type tailValue struct {
	Time     time.Time
	Messages []string
}

func (v *tailValue) UnmarshalJSON(in []byte) error {
	var raw []string
	if err := json.Unmarshal(in, &raw); err != nil {
		return errors.Wrap(err, "unmarshal tailValue")
	}
	if len(raw) < 2 {
		return errors.Errorf(`unexpected tailValue; got %#v, want ["time", "message", ...]`, raw)
	}
	ts, err := strconv.ParseInt(raw[0], 10, 64)
	if err != nil {
		return errors.Wrapf(err, "parse tailValue time %s into integer", raw[0])
	}
	v.Time = time.Unix(0, ts)
	v.Messages = raw[1:]
	return nil
}

// TailChunk is a set of log lines retrieved via Loki's tail API.  Each line occurred at the same
// time and has the same associated fields.
type TailChunk struct {
	Time   time.Time         // The time of the messages.
	Fields map[string]string // Any loki fields attached to the messages.
	Lines  []string          // The text of the messages.
}

// The websocket reader does not respect the context while reading, so we set an explicit deadline
// for each read.  It's OK if no log lines show up in this interval, the read will be silently
// retried if this deadline is reached and the context is still alive.  This is a variable so tests
// can set this to something short and end as soon as they are ready to end.
var TailPerReadDeadline = 30 * time.Second

// Tail watches a Loki query until the provided context is canceled.  The start time should be as
// close as possible to the present time; you can get some historical logs through this API, but
// logs will be missed while that query is running on the loki side.  So it's best to call
// QueryRange up to the current timestamp, and then pass the most recent timestamp that call
// received to Tail.  This function will always return an error; if you canceled the context on
// purpose you should ignore the error context.Canceled.
func (c *Client) Tail(ctx context.Context, start time.Time, query string, cb func(*TailChunk) error) error {
	var connected bool
	delay := backoff.NewExponentialBackOff()
main:
	for {
		// Calculate the URL to connect to.  This changes every iteration, because start
		// advances as we read logs.
		values := make(url.Values)
		values.Add("query", query)
		values.Add("start", strconv.FormatInt(start.UnixNano(), 10))
		u, err := buildURL(c.Address, "/loki/api/v1/tail", values.Encode())
		if err != nil {
			return errors.Wrap(err, "build tailing url")
		}
		if http := "http:"; strings.HasPrefix(u, http) {
			u = "ws:" + u[len(http):]
		}
		cfg, err := websocket.NewConfig(u, "ws://pachyderm-loki")
		if err != nil {
			return errors.Wrap(err, "build websocket config")
		}

		// Try connecting.
		log.Debug(ctx, "attempting to connect to loki", zap.String("url", u), zap.Bool("previously_connected", connected))
		conn, err := cfg.DialContext(ctx)
		if err != nil {
			if !connected {
				// The first connection will fail immediately.  Subsequent reconnect
				// attempts will just retry until the context is done.  (If it
				// worked once, then it's probably just a transient error on the
				// Loki side.)
				return errors.Wrap(err, "dial loki tail endpoint")
			}
			select {
			case <-time.After(delay.NextBackOff()):
				continue main
			case <-ctx.Done():
				return errors.Wrap(context.Cause(ctx), "context done while awaiting loki reconnect after dial failure")
			}
		}
		connected = true
		delay.Reset()
		log.Debug(ctx, "connected to loki websocket", zap.String("url", u))
		for {
			select {
			case <-ctx.Done():
				return errors.Wrap(context.Cause(ctx), "context done while awaiting loki websocket message")
			default:
			}

			// This read is not otherwise gated by the context, so we set a short
			// deadline on the read.  If the context is OK after the failed read, we
			// just restart.
			if err := conn.SetDeadline(time.Now().Add(TailPerReadDeadline)); err != nil {
				return errors.Wrap(err, "set websocket activity deadline")
			}
			var res tailResponse

			// Read a message.
			if err := websocket.JSON.Receive(conn, &res); err != nil {
				if errors.Is(err, io.EOF) {
					// The Loki server closed the connection.  It will probably
					// be back, so reconnect.
					select {
					case <-time.After(delay.NextBackOff()):
						continue main
					case <-ctx.Done():
						return errors.Wrap(context.Cause(ctx), "context done while awaiting loki reconnect after EOF")
					}
				}
				if errors.Is(err, os.ErrDeadlineExceeded) {
					// We didn't get a message within our read deadline; that is
					// expected.  The context is checked again at the top of the
					// loop.
					continue
				}
				return errors.Wrap(err, "decoding loki response")
			}

			// Order the messages in the response by time.  Loki chunks them into
			// related log streams, but we want to read in time order instead.
			var chunks []*TailChunk
			for _, stream := range res.Streams {
				for _, value := range stream.Values {
					if !value.Time.IsZero() && value.Time.After(start) {
						// Advance pointer for reconnect.
						start = value.Time
					}
					chunks = append(chunks, &TailChunk{
						Time:   value.Time,
						Fields: stream.Stream,
						Lines:  value.Messages,
					})
				}
			}
			sort.Slice(chunks, func(i, j int) bool {
				return chunks[i].Time.Before(chunks[j].Time)
			})
			for _, c := range chunks {
				if err := cb(c); err != nil {
					return errors.Wrap(err, "run log chunk handling callback")
				}
			}
		}
	}
}
