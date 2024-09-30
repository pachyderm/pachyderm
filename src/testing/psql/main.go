// Command psql connects to postgres-0 setup by pachdev's default settings in your current k8s
// namespace and starts a `psql` shell against it.  Any args to this command are passed directly to
// psql.
//
// This is a quick hack; it should be part of pachdev but pachdev needs to be refactored to not
// build images on every run.  It should then use your pachdev context / postgres password stored in
// the cluster / etc.
package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

const (
	forwardingFrom     = "Forwarding from 127.0.0.1:"
	portForwardStartup = 10 * time.Second
)

func main() {
	endLogging := log.InitBatchLogger("")
	ctx, cancel := pctx.Interactive()

	err := Run(ctx)
	cancel()
	if err != nil {
		log.Error(ctx, err.Error())
	}
	endLogging(err)
}

func Run(ctx context.Context) error {
	port, err := StartPortForward(ctx)
	if err != nil {
		return errors.Wrap(err, "starting port-forward")
	}
	if err := RunPsql(ctx, "127.0.0.1", port, "root"); err != nil {
		return errors.Wrap(err, "running psql")
	}
	return nil
}

// StartPortForward starts a `kubectl port-forward` run, returning as soon as port-forward tells us
// which random port it picked, or returning an error if port-forward fails to output a port number
// or fails to startup.
func StartPortForward(ctx context.Context) (_ int, retErr error) {
	ctx, done := log.SpanContext(ctx, "kubectl.port-forward")
	defer done(log.Errorp(&retErr))

	path, ok := bazel.FindBinary("//tools/kubectl", "_kubectl")
	if !ok {
		log.Info(ctx, "binary not built with bazel; falling back to host kubectl")
		path = "kubectl"
	}
	reportedError := new(atomic.Bool)
	r, w := io.Pipe()
	cmd := exec.CommandContext(ctx, path, "port-forward", "postgres-0", "0:5432", "--address=127.0.0.1")
	cmd.Stdin = nil
	cmd.Stdout = w
	cmd.Stderr = log.WriterAt(pctx.Child(ctx, "stderr"), log.InfoLevel)
	startupTimer := time.NewTimer(portForwardStartup)
	go func() {
		select {
		case <-startupTimer.C:
			w.CloseWithError(errors.New("startup deadline exceeded"))
		case <-ctx.Done():
			w.CloseWithError(context.Cause(ctx))
		}
	}()
	if err := cmd.Start(); err != nil {
		return 0, errors.Wrap(err, "start kubectl port-forward")
	}
	wait := func() {
		err := cmd.Wait()
		if err != nil && !reportedError.Load() {
			if err.Error() != "signal: killed" {
				log.Error(ctx, "port-forward exited unexpectedly", zap.Error(err))
			}
		}
		w.CloseWithError(errors.Wrap(err, "wait for port-forward to exit"))
	}
	go wait()

	// Wait for a line that looks like "Forwarding from 127.0.0.1:12345 -> 5432"
	s := bufio.NewScanner(r)
	var port int64
	for s.Scan() {
		line := s.Text()
		if !strings.HasPrefix(line, forwardingFrom) {
			log.Info(ctx, "stdout: "+line)
			continue
		}
		addr := line[len(forwardingFrom):]
		if i := strings.Index(addr, " ->"); i > 0 {
			addr = addr[:i]
		}
		var err error
		if port, err = strconv.ParseInt(addr, 10, 32); err != nil {
			log.Info(ctx, "problem parsing port from kubectl port-forward", zap.String("line", line), zap.Error(err))
			continue
		}
		break
	}
	startupTimer.Stop()

	if err := s.Err(); err != nil {
		if port == 0 {
			log.Error(ctx, "problem reading output from port-forward", zap.Error(err))
			reportedError.Store(true)
			return 0, errors.Wrap(err, "port-forward did not output a port number")
		} else {
			log.Info(ctx, "problem reading output from port-forward, but we got a port, so continuing...", zap.Error(err))
		}
	}
	go func() {
		for s.Scan() { // drain the reader
			log.Info(ctx, "stdout: "+s.Text())
		}
	}()
	return int(port), nil
}

// RunPsql runs `psql` in the foreground (so that readline, etc. work) and returns when it exits.
// os.Argv[1:] is passed to the psql command.
func RunPsql(ctx context.Context, host string, port int, password string) error {
	args := []string{"-U", "postgres", "-h", host, "-p", strconv.Itoa(port)}
	args = append(args, os.Args[1:]...)
	cmd := exec.CommandContext(ctx, "psql", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	environ := os.Environ()
	environ = append(environ, "PGPASSWORD="+password)
	cmd.Env = environ
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: true,
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "run psql")
	}
	return nil
}
