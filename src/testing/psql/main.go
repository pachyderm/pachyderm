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
	"time"

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
	log.SetLevel(log.DebugLevel)
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

	r, w := io.Pipe()
	cmd := exec.CommandContext(ctx, "kubectl", "port-forward", "postgres-0", "0:5432", "--address=127.0.0.1")
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

	// Wait for a line that looks like "Forwarding from 127.0.0.1:12345 -> 5432"
	s := bufio.NewScanner(r)
	var port int64
	for s.Scan() {
		line := s.Text()
		if port > 0 {
			log.Info(ctx, line)
			continue
		}
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
		// This sort of error would probably prevent us from starting up but not from
		// continuing the run.  If it breaks the run, Wait() below will abort.
		if port > 0 {
			log.Info(ctx, "problem reading output from port-forward, but it doesn't really matter", zap.Error(err))
		} else {
			log.Error(ctx, "problem reading output from port-forward", zap.Error(err))
		}
	}
	go func() {
		for s.Scan() { // drain the reader
			log.Info(ctx, s.Text())
		}
	}()

	// If we didn't read a port number, kill port-forward and wait for port-forward to exit.
	wait := func() error {
		if err := cmd.Wait(); err != nil {
			if err.Error() != "signal: killed" {
				return errors.Wrap(err, "wait for port-forward to exit")
			}
		}
		return nil
	}
	if port == 0 {
		if err := cmd.Process.Kill(); err != nil {
			return 0, errors.Wrap(err, "kill zombie port-forward")
		}
		if err := wait(); err != nil {
			return 0, errors.Wrap(err, "after never outputting a port number")
		}
		return 0, errors.New("port-forward did not output a port number")
	}

	// We read a port number, so wait for port-forward to exit (when the context is canceled) in
	// the background.
	go func() {
		if err := wait(); err != nil {
			log.Error(ctx, "problem running port-forward", zap.Error(err))
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
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "run psql")
	}
	return nil
}
