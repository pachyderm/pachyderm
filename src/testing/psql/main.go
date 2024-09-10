// Command psql connects to postgres-0 in your current k8s namespace and starts a `psql` shell
// against it.
//
// This is a quick hack; it should be part of pachdev but pachdev needs to be refactored to not
// build images on every run.  It should then use your pachdev context / postgres password / etc.
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
		log.Error(ctx, "run failed", zap.Error(err))
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

func StartPortForward(ctx context.Context) (_ int, retErr error) {
	ctx, done := log.SpanContext(ctx, "kubectl.port-forward")
	defer done(log.Errorp(&retErr))

	cmd := exec.CommandContext(ctx, "kubectl", "port-forward", "postgres-0", "0:5432", "--address=127.0.0.1")
	cmd.Stdin = nil
	r, w := io.Pipe()

	startupTimer := time.NewTimer(portForwardStartup)
	go func() {
		select {
		case <-startupTimer.C:
			w.CloseWithError(errors.New("startup deadline exceeded"))
		case <-ctx.Done():
			w.CloseWithError(context.Cause(ctx))
		}
	}()
	cmd.Stdout = w
	cmd.Stderr = log.WriterAt(pctx.Child(ctx, "stderr"), log.InfoLevel)
	if err := cmd.Start(); err != nil {
		return 0, errors.Wrap(err, "start kubectl port-forward")
	}
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
	go func() {
		if err := cmd.Wait(); err != nil {
			if err.Error() != "signal: killed" {
				log.Error(ctx, "port-forward ended", zap.Error(err))
			}
			return
		}
		log.Info(ctx, "port-forward ended")
	}()
	return int(port), nil
}

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
