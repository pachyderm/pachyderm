package client_test

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/testloki"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestMain(m *testing.M) {
	client.TailPerReadDeadline = 100 * time.Millisecond
	m.Run()
}

func TestTail(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	loki, err := testloki.New(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("set up loki: %v", err)
	}
	t.Cleanup(func() {
		if err := loki.Close(); err != nil {
			t.Fatalf("clean up loki: %v", err)
		}
	})
	var line int
	want := []*client.TailChunk{
		{
			Fields: map[string]string{"app": "a", "suite": "pachyderm"},
			Lines:  []string{"starting 1"},
		},
		{
			Fields: map[string]string{"app": "b", "suite": "pachyderm"},
			Lines:  []string{"starting 2"},
		},
		{
			Fields: map[string]string{"app": "a", "suite": "pachyderm"},
			Lines:  []string{"line 3"},
		},
		{
			Fields: map[string]string{"app": "b", "suite": "pachyderm"},
			Lines:  []string{"line 4"},
		},
		{
			Fields: map[string]string{"app": "a", "suite": "pachyderm"},
			Lines:  []string{"line 5"},
		},
		{
			Fields: map[string]string{"app": "b", "suite": "pachyderm"},
			Lines:  []string{"line 6"},
		},
		{
			Fields: map[string]string{"app": "a", "suite": "pachyderm"},
			Lines:  []string{"line 7"},
		},
		{
			Fields: map[string]string{"app": "b", "suite": "pachyderm"},
			Lines:  []string{"line 8"},
		},
		{
			Fields: map[string]string{"app": "a", "suite": "pachyderm"},
			Lines:  []string{"line 9"},
		},
		{
			Fields: map[string]string{"app": "b", "suite": "pachyderm"},
			Lines:  []string{"line 10"},
		},
	}
	send := func(app, msg string) {
		ts := time.Now()
		if err := loki.AddLog(ctx, &testloki.Log{
			Time:    ts,
			Message: msg,
			Labels:  map[string]string{"app": app, "suite": "pachyderm"},
		}); err != nil {
			t.Fatalf("add log: %v", err)
		}
		// Loki will be mad if logs are too far in the past, so we have to use time.Now() or
		// similar.  This adds the time we sent into the expected return value (with the
		// time.Unix() thing to strip off the monotonic part of the timestamp).
		want[line].Time = time.Unix(0, ts.UnixNano())
		line++
	}
	send("a", "starting 1")
	send("b", "starting 2")

	tctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	errDone := errors.New("done")
	var got []*client.TailChunk
	var gotOne bool
	cb := func(chunk *client.TailChunk) error {
		got = append(got, chunk)
		if line >= 10 {
			t.Log("recevied enough lines; ending tail")
			return errDone
		}
		app := "a"
		if line%2 == 1 {
			app = "b"
		}
		if gotOne {
			send(app, fmt.Sprintf("line %v", line+1))
		} else {
			// This keeps the lines reading logically; line 1 arrives, we do nothing,
			// line 2 arrives, we send line 3, etc.
			gotOne = true
		}
		return nil
	}
	if err := loki.Client.Tail(tctx, time.Now().Add(-24*time.Hour), `{suite="pachyderm"}`, cb); err != nil {
		if !errors.Is(err, errDone) {
			t.Fatalf("tail ended in error: %v", err)
		}
	}
	require.NoDiff(t, want, got, nil, "received lines should match expected lines")
}

func TestTail_BadConnection(t *testing.T) {
	t.Parallel()
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("find port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		t.Fatalf("free listener: %v", err)
	}
	c := client.Client{Address: fmt.Sprintf("http://127.0.0.1:%d", port)}

	ctx := pctx.TestContext(t)
	if err := c.Tail(ctx, time.Now(), "{}", func(tc *client.TailChunk) error {
		t.Fatalf("callback should not have been called (with %#v)", tc)
		return nil
	}); err == nil {
		t.Error("expected an error, but got success")
	} else if got, want := err.Error(), "dial loki tail"; !strings.Contains(got, want) {
		t.Errorf("unexpected error:\n  got: %v\n want: %v", got, want)
	}
}

func TestTail_LokiDies(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	loki, err := testloki.New(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("set up loki: %v", err)
	}
	t.Cleanup(func() {
		if err := loki.Close(); err != nil {
			t.Fatalf("clean up loki: %v", err)
		}
	})
	if err := loki.AddLog(ctx, &testloki.Log{
		Time:    time.Now(),
		Message: "hello",
		Labels:  map[string]string{"app": "foo"},
	}); err != nil {
		t.Fatalf("add log: %v", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)
	var gotLog bool
	if err := loki.Client.Tail(ctx, time.Now().Add(-time.Minute), `{app="foo"}`, func(tc *client.TailChunk) error {
		gotLog = true
		if err := loki.Close(); err != nil {
			return errors.Wrap(err, "close loki")
		}
		return nil
	}); err == nil {
		t.Error("expected an error but got success")
	} else if strings.Contains(err.Error(), "close loki") {
		t.Errorf("problem killing loki: %v", err)
	} else {
		t.Logf("tail ended (this is expected): %v", err)
	}
	if !gotLog {
		t.Error("never got the log that would have triggered loki to exit")
	}
}
