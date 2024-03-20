package testloki

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// TestLoki is a real Loki instance.
type TestLoki struct {
	Port      int
	Client    *client.Client
	killLoki  func(error)
	lokiErrCh <-chan error
}

// New starts a new Loki instance on the local machine.
func New(ctx context.Context, tmp string) (*TestLoki, error) {
	bin, ok := bazel.FindBinary("//tools/loki", "loki")
	if !ok {
		log.Debug(ctx, "can't find //tools/loki via bazel, using loki in $PATH")
		bin = "loki"
	}

	templateFile, err := runfiles.Rlocation("_main/src/internal/lokiutil/testloki/config.yaml")
	if err != nil {
		log.Debug(ctx, "warning: can't find config.yaml via bazel, using ./config.yaml")
		templateFile = "config.yaml"
	}
	template, err := os.ReadFile(templateFile)
	if err != nil {
		return nil, errors.Wrap(err, "read config.yaml template")
	}

	l, err := net.Listen("tcp", "")
	if err != nil {
		return nil, errors.Wrap(err, "find random port: listen")
	}
	if err := l.Close(); err != nil {
		return nil, errors.Wrap(err, "find random port: close")
	}
	port := l.Addr().(*net.TCPAddr).Port

	chunks, rules := filepath.Join(tmp, "chunks"), filepath.Join(tmp, "rules")
	if err := os.MkdirAll(chunks, 0o755); err != nil {
		return nil, errors.Wrap(err, "make chunk dir")
	}
	if err := os.MkdirAll(chunks+"-temp", 0o755); err != nil {
		// loki logs this being missing at level "error", but it's crash recovery code, so
		// it shouldn't exist.  sigh.
		return nil, errors.Wrap(err, "make chunk-temp dir")
	}
	if err := os.MkdirAll(rules, 0o755); err != nil {
		return nil, errors.Wrap(err, "make rules dir")
	}
	if err := os.MkdirAll(rules+"-temp", 0o755); err != nil {
		return nil, errors.Wrap(err, "make rules-temp dir")
	}

	var config map[string]any
	if err := yaml.Unmarshal(template, &config); err != nil {
		return nil, errors.Wrapf(err, "unmarshal config.yaml template from %v", template)
	}
	config["common"].(map[string]any)["path_prefix"] = tmp
	config["common"].(map[string]any)["storage"] = map[string]any{
		"filesystem": map[string]any{
			"chunks_directory": chunks,
			"rules_directory":  rules,
		},
	}
	config["server"].(map[string]any)["http_listen_port"] = port
	config["server"].(map[string]any)["grpc_listen_port"] = port + 1
	config["schema_config"].(map[string]any)["configs"].([]any)[0].(map[string]any)["from"] = "2020-10-24"

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return nil, errors.Wrap(err, "marshal config.yaml")
	}

	configFile := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(configFile, configBytes, 0o600); err != nil {
		return nil, errors.Wrap(err, "write config.yaml")
	}

	ctx, killLoki := context.WithCancelCause(ctx)
	lokiErrCh := make(chan error)
	loki := exec.CommandContext(ctx, bin, "-config.file", configFile, "-log.level", "error")
	loki.Stdout = io.Discard
	loki.Stderr = log.WriterAt(pctx.Child(ctx, "loki.stderr"), log.DebugLevel)
	if err := loki.Start(); err != nil {
		return nil, errors.Wrap(err, "start loki")
	}
	go func() {
		if err := loki.Wait(); err != nil {
			exitErr := new(exec.ExitError)
			if !errors.As(err, &exitErr) {
				lokiErrCh <- errors.Wrap(err, "wait for loki to exit")
			}
		}
		close(lokiErrCh)
		// Avoid running the polling code below if Loki immediately exits.  This
		// happens if the config is broken.
		killLoki(errors.New("loki exited"))
	}()

	addr := "http://127.0.0.1:" + strconv.Itoa(port)
	for i := 0; i < 300; i++ {
		select {
		case <-ctx.Done():
			return nil, errors.Wrap(context.Cause(ctx), "context done")
		default:
		}
		// Loki, with our config, starts up in about 3 seconds.  If it isn't ready after 5s,
		// then start logging every 10th attempt.
		doLog := i > 50 && i%10 == 0
		err := pingLoki(ctx, addr, doLog)
		if err == nil {
			return &TestLoki{
				Port: port,
				Client: &client.Client{
					Address: addr,
				},
				killLoki:  killLoki,
				lokiErrCh: lokiErrCh,
			}, nil
		}
		if doLog {
			log.Debug(ctx, "loki not ready; retrying in 100ms", zap.Error(err))
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, errors.New("loki failed to start up after 30s")
}

func pingLoki(ctx context.Context, addr string, doLog bool) (retErr error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", addr+"/ready", nil)
	if err != nil {
		return errors.Wrap(err, "build /ready request")
	}
	client := &http.Client{
		Transport: promutil.InstrumentRoundTripper("testloki.ping", nil),
	}
	if !doLog {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "do request")
	}
	defer errors.Close(&retErr, res.Body, "close body")
	if got, want := res.StatusCode, http.StatusOK; got != want {
		body, _ := io.ReadAll(res.Body) // if this errors, fine, no extra info in error below
		return errors.Errorf("loki not ready (%v): %s", res.Status, body)
	}
	return nil
}

// Close kills the running Loki.
func (l *TestLoki) Close() error {
	l.killLoki(errors.New("killed with TestLoki.Close()"))
	if err := <-l.lokiErrCh; err != nil {
		return errors.Wrap(err, "wait for loki to exit")
	}
	return nil
}

// Log is a log entry to add to Loki.
type Log struct {
	Time    time.Time
	Message string
	Labels  map[string]string
}

type pushRequest struct {
	Streams []stream `json:"streams"`
}

type stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// AddLog adds a log line to Loki.
func (l *TestLoki) AddLog(ctx context.Context, lg *Log) (retErr error) {
	pr := &pushRequest{
		Streams: []stream{
			{
				Stream: lg.Labels,
				Values: [][]string{
					{
						strconv.FormatInt(lg.Time.UnixNano(), 10),
						lg.Message,
					},
				},
			},
		},
	}
	js, err := json.Marshal(pr)
	if err != nil {
		return errors.Wrapf(err, "marshal request %#v", pr)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", l.Client.Address+"/loki/api/v1/push", bytes.NewReader(js))
	if err != nil {
		return errors.Wrap(err, "new request")
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", "testloki")

	client := &http.Client{
		Transport: promutil.InstrumentRoundTripper("testloki.ping", nil),
	}
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "do request")
	}
	defer errors.Close(&retErr, res.Body, "close body")
	if res.StatusCode/100 != 2 {
		body, _ := io.ReadAll(res.Body)
		return errors.Wrapf(err, "unexpected response status %v: %v", res.StatusCode, body)
	}
	return nil
}
