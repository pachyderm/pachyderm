package testloki

import (
	"context"
	"io"
	"net"
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
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// TestLoki is a real Loki instance.
type TestLoki struct {
	Port    int
	Client  *client.Client
	LokiCmd *exec.Cmd
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

	var config map[string]any
	if err := yaml.Unmarshal(template, &config); err != nil {
		return nil, errors.Wrapf(err, "unmarshal config.yaml template from %v", template)
	}

	config["common"].(map[string]any)["path_prefix"] = tmp
	config["common"].(map[string]any)["storage"] = map[string]any{
		"filesystem": map[string]any{
			"chunks_directory": filepath.Join(tmp, "chunks"),
			"rules_directory":  filepath.Join(tmp, "rules"),
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

	loki := exec.CommandContext(ctx, bin, "-config.file", configFile, "-log.level", "error")
	loki.Stdout = io.Discard
	loki.Stderr = log.WriterAt(pctx.Child(ctx, "loki.stderr"), log.DebugLevel)
	if err := loki.Start(); err != nil {
		return nil, errors.Wrap(err, "start loki")
	}

	addr := "http://127.0.0.1:" + strconv.Itoa(port)
	c := &client.Client{
		Address: addr,
	}
	for {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		_, err := c.QueryRange(ctx, `{host=~".+"}`, 1000, time.Now(), time.Time{}, "BACKWARD", time.Second, time.Second, false)
		cancel()
		if err == nil {
			return &TestLoki{
				Port:    port,
				Client:  c,
				LokiCmd: loki,
			}, nil
		}
		log.Debug(ctx, "loki not ready; retrying in 100ms", zap.Error(err))
		time.Sleep(100 * time.Millisecond)
	}
}

// Close kills the running Loki.
func (l *TestLoki) Close() error {
	if l.LokiCmd == nil {
		return errors.New("already closed")
	}
	if err := l.LokiCmd.Process.Kill(); err != nil {
		return errors.Wrap(err, "kill loki")
	}
	if err := l.LokiCmd.Wait(); err != nil {
		exitErr := new(exec.ExitError)
		if !errors.As(err, &exitErr) {
			return errors.Wrap(err, "wait for loki to exit")
		}
	}
	l.LokiCmd = nil
	return nil
}
