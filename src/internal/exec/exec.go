// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package exec runs external commands. It wraps os.StartProcess to make it
// easier to remap stdin and stdout, connect I/O with pipes, and do other
// adjustments.
//
// Unlike the "system" library call from C and other languages, the
// os/exec package intentionally does not invoke the system shell and
// does not expand any glob patterns or handle other expansions,
// pipelines, or redirections typically done by shells. The package
// behaves more like C's "exec" family of functions. To expand glob
// patterns, either call the shell directly, taking care to escape any
// dangerous input, or use the path/filepath package's Glob function.
// To expand environment variables, use package os's ExpandEnv.
//
// Note that the examples in this package assume a Unix system.
// They may not run on Windows, and they do not run in the Go Playground
// used by golang.org and godoc.org.
package exec

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// Error records the name of a binary that failed to be executed
// and the reason it failed.
type Error struct {
	Name string
	Err  error
}

func (e *Error) Error() string {
	return "exec: " + strconv.Quote(e.Name) + ": " + e.Err.Error()
}

// Cmd represents an external command being prepared or run.
//
// A Cmd cannot be reused after calling its Run, Output or CombinedOutput
// methods.
type Cmd struct {
	// Path is the path of the command to run.
	//
	// This is the only field that must be set to a non-zero
	// value. If Path is relative, it is evaluated relative
	// to Dir.
	Path string

	// Args holds command line arguments, including the command as Args[0].
	// If the Args field is empty or nil, Run uses {Path}.
	//
	// In typical use, both Path and Args are set by calling Command.
	Args []string

	// Env specifies the environment of the process.
	// Each entry is of the form "key=value".
	// If Env is nil, the new process uses the current process's
	// environment.
	// If Env contains duplicate environment keys, only the last
	// value in the slice for each duplicate key is used.
	Env []string

	// Dir specifies the working directory of the command.
	// If Dir is the empty string, Run runs the command in the
	// calling process's current directory.
	Dir string

	// Stdin specifies the process's standard input.
	// If Stdin is nil, the process reads from the null device (os.DevNull).
	// If Stdin is an *os.File, the process's standard input is connected
	// directly to that file.
	// Otherwise, during the execution of the command a separate
	// goroutine reads from Stdin and delivers that data to the command
	// over a pipe. In this case, Wait does not complete until the goroutine
	// stops copying, either because it has reached the end of Stdin
	// (EOF or a read error) or because writing to the pipe returned an error.
	Stdin io.Reader

	// Stdout and Stderr specify the process's standard output and error.
	//
	// If either is nil, Run connects the corresponding file descriptor
	// to the null device (os.DevNull).
	//
	// If Stdout and Stderr are the same writer, and have a type that can be compared with ==,
	// at most one goroutine at a time will call Write.
	Stdout io.Writer
	Stderr io.Writer

	// ExtraFiles specifies additional open files to be inherited by the
	// new process. It does not include standard input, standard output, or
	// standard error. If non-nil, entry i becomes file descriptor 3+i.
	ExtraFiles []*os.File

	// SysProcAttr holds optional, operating system-specific attributes.
	// Run passes it to os.StartProcess as the os.ProcAttr's Sys field.
	SysProcAttr *syscall.SysProcAttr

	// Process is the underlying process, once started.
	Process *os.Process

	// ProcessState contains information about an exited process,
	// available after a call to Wait or Run.
	ProcessState *os.ProcessState

	ctx             context.Context // nil means none
	lookPathErr     error           // LookPath error, if any.
	finished        bool            // when Wait was called
	childFiles      []*os.File
	closeAfterStart []io.Closer
	closeAfterWait  []io.Closer
	goroutine       []func() error
	errch           chan error // one send per goroutine
	waitDone        chan struct{}
}

// Command returns the Cmd struct to execute the named program with
// the given arguments.
//
// It sets only the Path and Args in the returned structure.
//
// If name contains no path separators, Command uses LookPath to
// resolve name to a complete path if possible. Otherwise it uses name
// directly as Path.
//
// The returned Cmd's Args field is constructed from the command name
// followed by the elements of arg, so arg should not include the
// command name itself. For example, Command("echo", "hello").
// Args[0] is always name, not the possibly resolved Path.
func Command(name string, arg ...string) *Cmd {
	cmd := &Cmd{
		Path: name,
		Args: append([]string{name}, arg...),
	}
	if filepath.Base(name) == name {
		if lp, err := exec.LookPath(name); err != nil {
			cmd.lookPathErr = err
		} else {
			cmd.Path = lp
		}
	}
	return cmd
}

// CommandContext is like Command but includes a context.
//
// The provided context is used to kill the process (by calling
// os.Process.Kill) if the context becomes done before the command
// completes on its own.
func CommandContext(ctx context.Context, name string, arg ...string) *Cmd {
	if ctx == nil {
		panic("nil Context")
	}
	cmd := Command(name, arg...)
	cmd.ctx = ctx
	return cmd
}

// interfaceEqual protects against panics from doing equality tests on
// two interfaces with non-comparable underlying types.
func interfaceEqual(a, b interface{}) bool {
	defer func() {
		recover() //nolint:errcheck
	}()
	return a == b
}

func (c *Cmd) envv() []string {
	if c.Env != nil {
		return c.Env
	}
	return os.Environ()
}

func (c *Cmd) argv() []string {
	if len(c.Args) > 0 {
		return c.Args
	}
	return []string{c.Path}
}

// skipStdinCopyError optionally specifies a function which reports
// whether the provided the stdin copy error should be ignored.
// It is non-nil everywhere but Plan 9, which lacks EPIPE. See exec_posix.go.
var skipStdinCopyError func(error) bool

func (c *Cmd) stdin() (f *os.File, err error) {
	if c.Stdin == nil {
		f, err = os.Open(os.DevNull)
		if err != nil {
			return
		}
		c.closeAfterStart = append(c.closeAfterStart, f)
		return
	}

	if f, ok := c.Stdin.(*os.File); ok {
		return f, nil
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		return
	}

	c.closeAfterStart = append(c.closeAfterStart, pr)
	c.closeAfterWait = append(c.closeAfterWait, pw)
	c.goroutine = append(c.goroutine, func() error {
		_, err := io.Copy(pw, c.Stdin)
		if skip := skipStdinCopyError; skip != nil && skip(err) {
			err = nil
		}
		if err1 := pw.Close(); err == nil {
			err = err1
		}
		return err
	})
	return pr, nil
}

func (c *Cmd) stdout() (f *os.File, err error) {
	return c.writerDescriptor(c.Stdout)
}

func (c *Cmd) stderr() (f *os.File, err error) {
	if c.Stderr != nil && interfaceEqual(c.Stderr, c.Stdout) {
		return c.childFiles[1], nil
	}
	return c.writerDescriptor(c.Stderr)
}

func (c *Cmd) writerDescriptor(w io.Writer) (f *os.File, err error) {
	if w == nil {
		f, err = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err != nil {
			return
		}
		c.closeAfterStart = append(c.closeAfterStart, f)
		return
	}

	if f, ok := w.(*os.File); ok {
		return f, nil
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		return
	}

	c.closeAfterStart = append(c.closeAfterStart, pw)
	c.closeAfterWait = append(c.closeAfterWait, pr)
	c.goroutine = append(c.goroutine, func() error {
		_, err := io.Copy(w, pr)
		pr.Close() // in case io.Copy stopped due to write error
		return errors.EnsureStack(err)
	})
	return pw, nil
}

func (c *Cmd) closeDescriptors(closers []io.Closer) {
	for _, fd := range closers {
		fd.Close()
	}
}

// Run starts the specified command and waits for it to complete.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
//
// If the command starts but does not complete successfully, the error is of
// type *ExitError. Other error types may be returned for other situations.
func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

// lookExtensions finds windows executable by its dir and path.
// It uses LookPath to try appropriate extensions.
// lookExtensions does not search PATH, instead it converts `prog` into `.\prog`.
func lookExtensions(path, dir string) (string, error) {
	if filepath.Base(path) == path {
		path = filepath.Join(".", path)
	}
	if dir == "" {
		res, err := exec.LookPath(path)
		return res, errors.EnsureStack(err)
	}
	if filepath.VolumeName(path) != "" {
		res, err := exec.LookPath(path)
		return res, errors.EnsureStack(err)
	}
	if len(path) > 1 && os.IsPathSeparator(path[0]) {
		res, err := exec.LookPath(path)
		return res, errors.EnsureStack(err)
	}
	dirandpath := filepath.Join(dir, path)
	// We assume that LookPath will only add file extension.
	lp, err := exec.LookPath(dirandpath)
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	ext := strings.TrimPrefix(lp, dirandpath)
	return path + ext, nil
}

// Start starts the specified command but does not wait for it to complete.
//
// The Wait method will return the exit code and release associated resources
// once the command exits.
func (c *Cmd) Start() error {
	if c.lookPathErr != nil {
		c.closeDescriptors(c.closeAfterStart)
		c.closeDescriptors(c.closeAfterWait)
		return c.lookPathErr
	}
	if runtime.GOOS == "windows" {
		lp, err := lookExtensions(c.Path, c.Dir)
		if err != nil {
			c.closeDescriptors(c.closeAfterStart)
			c.closeDescriptors(c.closeAfterWait)
			return err
		}
		c.Path = lp
	}
	if c.Process != nil {
		return errors.New("exec: already started")
	}
	if c.ctx != nil {
		select {
		case <-c.ctx.Done():
			c.closeDescriptors(c.closeAfterStart)
			c.closeDescriptors(c.closeAfterWait)
			return errors.EnsureStack(c.ctx.Err())
		default:
		}
	}

	type F func(*Cmd) (*os.File, error)
	for _, setupFd := range []F{(*Cmd).stdin, (*Cmd).stdout, (*Cmd).stderr} {
		fd, err := setupFd(c)
		if err != nil {
			c.closeDescriptors(c.closeAfterStart)
			c.closeDescriptors(c.closeAfterWait)
			return err
		}
		c.childFiles = append(c.childFiles, fd)
	}
	c.childFiles = append(c.childFiles, c.ExtraFiles...)

	var err error
	c.Process, err = os.StartProcess(c.Path, c.argv(), &os.ProcAttr{
		Dir:   c.Dir,
		Files: c.childFiles,
		Env:   dedupEnv(c.envv()),
		Sys:   c.SysProcAttr,
	})
	if err != nil {
		c.closeDescriptors(c.closeAfterStart)
		c.closeDescriptors(c.closeAfterWait)
		return errors.EnsureStack(err)
	}

	c.closeDescriptors(c.closeAfterStart)

	c.errch = make(chan error, len(c.goroutine))
	for _, fn := range c.goroutine {
		go func(fn func() error) {
			c.errch <- fn()
		}(fn)
	}

	if c.ctx != nil {
		c.waitDone = make(chan struct{})
		go func() {
			select {
			case <-c.ctx.Done():
				c.Process.Kill() //nolint:errcheck
			case <-c.waitDone:
			}
		}()
	}

	return nil
}

// An ExitError reports an unsuccessful exit by a command.
type ExitError struct {
	*os.ProcessState

	// Stderr holds a subset of the standard error output from the
	// Cmd.Output method if standard error was not otherwise being
	// collected.
	//
	// If the error output is long, Stderr may contain only a prefix
	// and suffix of the output, with the middle replaced with
	// text about the number of omitted bytes.
	//
	// Stderr is provided for debugging, for inclusion in error messages.
	// Users with other needs should redirect Cmd.Stderr as needed.
	Stderr []byte
}

func (e *ExitError) Error() string {
	return e.ProcessState.String()
}

// Wait waits for the command to exit and waits for any copying to
// stdin or copying from stdout or stderr to complete.
//
// The command must have been started by Start.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
//
// If the command fails to run or doesn't complete successfully, the
// error is of type *ExitError. Other error types may be
// returned for I/O problems.
//
// If c.Stdin is not an *os.File, Wait also waits for the I/O loop
// copying from c.Stdin into the process's standard input
// to complete.
//
// Wait releases any resources associated with the Cmd.
func (c *Cmd) Wait() error {
	if c.Process == nil {
		return errors.New("exec: not started")
	}
	if c.finished {
		return errors.New("exec: Wait was already called")
	}
	c.finished = true

	state, err := c.Process.Wait()

	return c.WaitIO(state, err)
}

// WaitIO is a helper function and the reason we forked this
// package from stdlib. This way, we can manually close IO when
// c.Process.Wait() has already been called once
func (c *Cmd) WaitIO(state *os.ProcessState, err error) (retErr error) {
	if c.waitDone != nil {
		close(c.waitDone)
	}
	c.ProcessState = state
	if err != nil {
		return err
	} else if !state.Success() {
		return &ExitError{ProcessState: state}
	}
	for range c.goroutine {
		if err := <-c.errch; err != nil && retErr == nil {
			retErr = err
		}
	}
	c.closeDescriptors(c.closeAfterWait)
	return retErr
}

// Output runs the command and returns its standard output.
// Any returned error will usually be of type *ExitError.
// If c.Stderr was nil, Output populates ExitError.Stderr.
func (c *Cmd) Output() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	var stdout bytes.Buffer
	c.Stdout = &stdout

	captureErr := c.Stderr == nil
	if captureErr {
		c.Stderr = &prefixSuffixSaver{N: 32 << 10}
	}

	err := c.Run()
	if err != nil && captureErr {
		ee := &ExitError{}
		if errors.As(err, &ee) {
			ee.Stderr = c.Stderr.(*prefixSuffixSaver).Bytes()
		}
	}
	return stdout.Bytes(), err
}

// CombinedOutput runs the command and returns its combined standard
// output and standard error.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	err := c.Run()
	return b.Bytes(), err
}

// StdinPipe returns a pipe that will be connected to the command's
// standard input when the command starts.
// The pipe will be closed automatically after Wait sees the command exit.
// A caller need only call Close to force the pipe to close sooner.
// For example, if the command being run will not exit until standard input
// is closed, the caller must close the pipe.
func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	if c.Stdin != nil {
		return nil, errors.New("exec: Stdin already set")
	}
	if c.Process != nil {
		return nil, errors.New("exec: StdinPipe after process started")
	}
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	c.Stdin = pr
	c.closeAfterStart = append(c.closeAfterStart, pr)
	wc := &closeOnce{File: pw}
	c.closeAfterWait = append(c.closeAfterWait, closerFunc(wc.safeClose))
	return wc, nil
}

type closeOnce struct {
	*os.File

	writers sync.RWMutex // coordinate safeClose and Write
	once    sync.Once
	err     error
}

func (c *closeOnce) Close() error {
	c.once.Do(c.close)
	return c.err
}

func (c *closeOnce) close() {
	c.err = c.File.Close()
}

type closerFunc func() error

func (f closerFunc) Close() error { return f() }

// safeClose closes c being careful not to race with any calls to c.Write.
// See golang.org/issue/9307 and TestEchoFileRace in exec_test.go.
// In theory other calls could also be excluded (by writing appropriate
// wrappers like c.Write's implementation below), but since c is most
// commonly used as a WriteCloser, Write is the main one to worry about.
// See also #7970, for which this is a partial fix for this specific instance.
// The idea is that we return a WriteCloser, and so the caller can be
// relied upon not to call Write and Close simultaneously, but it's less
// obvious that cmd.Wait calls Close and that the caller must not call
// Write and cmd.Wait simultaneously. In fact that seems too onerous.
// So we change the use of Close in cmd.Wait to use safeClose, which will
// synchronize with any Write.
//
// It's important that we know this won't block forever waiting for the
// operations being excluded. At the point where this is called,
// the invoked command has exited and the parent copy of the read side
// of the pipe has also been closed, so there should really be no read side
// of the pipe left. Any active writes should return very shortly with an EPIPE,
// making it reasonable to wait for them.
// Technically it is possible that the child forked a sub-process or otherwise
// handed off the read side of the pipe before exiting and the current holder
// is not reading from the pipe, and the pipe is full, in which case the close here
// might block waiting for the write to complete. That's probably OK.
// It's a small enough problem to be outweighed by eliminating the race here.
func (c *closeOnce) safeClose() error {
	c.writers.Lock()
	err := c.Close()
	c.writers.Unlock()
	return err
}

func (c *closeOnce) Write(b []byte) (int, error) {
	c.writers.RLock()
	n, err := c.File.Write(b)
	c.writers.RUnlock()
	return n, errors.EnsureStack(err)
}

func (c *closeOnce) WriteString(s string) (int, error) {
	c.writers.RLock()
	n, err := c.File.WriteString(s)
	c.writers.RUnlock()
	return n, errors.EnsureStack(err)
}

// StdoutPipe returns a pipe that will be connected to the command's
// standard output when the command starts.
//
// Wait will close the pipe after seeing the command exit, so most callers
// need not close the pipe themselves; however, an implication is that
// it is incorrect to call Wait before all reads from the pipe have completed.
// For the same reason, it is incorrect to call Run when using StdoutPipe.
// See the example for idiomatic usage.
func (c *Cmd) StdoutPipe() (io.ReadCloser, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.Process != nil {
		return nil, errors.New("exec: StdoutPipe after process started")
	}
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	c.Stdout = pw
	c.closeAfterStart = append(c.closeAfterStart, pw)
	c.closeAfterWait = append(c.closeAfterWait, pr)
	return pr, nil
}

// StderrPipe returns a pipe that will be connected to the command's
// standard error when the command starts.
//
// Wait will close the pipe after seeing the command exit, so most callers
// need not close the pipe themselves; however, an implication is that
// it is incorrect to call Wait before all reads from the pipe have completed.
// For the same reason, it is incorrect to use Run when using StderrPipe.
// See the StdoutPipe example for idiomatic usage.
func (c *Cmd) StderrPipe() (io.ReadCloser, error) {
	if c.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	if c.Process != nil {
		return nil, errors.New("exec: StderrPipe after process started")
	}
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	c.Stderr = pw
	c.closeAfterStart = append(c.closeAfterStart, pw)
	c.closeAfterWait = append(c.closeAfterWait, pr)
	return pr, nil
}

// prefixSuffixSaver is an io.Writer which retains the first N bytes
// and the last N bytes written to it. The Bytes() methods reconstructs
// it with a pretty error message.
type prefixSuffixSaver struct {
	N         int // max size of prefix or suffix
	prefix    []byte
	suffix    []byte // ring buffer once len(suffix) == N
	suffixOff int    // offset to write into suffix
	skipped   int64

	// TODO(bradfitz): we could keep one large []byte and use part of it for
	// the prefix, reserve space for the '... Omitting N bytes ...' message,
	// then the ring buffer suffix, and just rearrange the ring buffer
	// suffix when Bytes() is called, but it doesn't seem worth it for
	// now just for error messages. It's only ~64KB anyway.
}

func (w *prefixSuffixSaver) Write(p []byte) (n int, err error) {
	lenp := len(p)
	p = w.fill(&w.prefix, p)

	// Only keep the last w.N bytes of suffix data.
	if overage := len(p) - w.N; overage > 0 {
		p = p[overage:]
		w.skipped += int64(overage)
	}
	p = w.fill(&w.suffix, p)

	// w.suffix is full now if p is non-empty. Overwrite it in a circle.
	for len(p) > 0 { // 0, 1, or 2 iterations.
		n := copy(w.suffix[w.suffixOff:], p)
		p = p[n:]
		w.skipped += int64(n)
		w.suffixOff += n
		if w.suffixOff == w.N {
			w.suffixOff = 0
		}
	}
	return lenp, nil
}

// fill appends up to len(p) bytes of p to *dst, such that *dst does not
// grow larger than w.N. It returns the un-appended suffix of p.
func (w *prefixSuffixSaver) fill(dst *[]byte, p []byte) (pRemain []byte) {
	if remain := w.N - len(*dst); remain > 0 {
		add := minInt(len(p), remain)
		*dst = append(*dst, p[:add]...)
		p = p[add:]
	}
	return p
}

func (w *prefixSuffixSaver) Bytes() []byte {
	if w.suffix == nil {
		return w.prefix
	}
	if w.skipped == 0 {
		return append(w.prefix, w.suffix...)
	}
	var buf bytes.Buffer
	buf.Grow(len(w.prefix) + len(w.suffix) + 50)
	buf.Write(w.prefix)
	buf.WriteString("\n... omitting ")
	buf.WriteString(strconv.FormatInt(w.skipped, 10))
	buf.WriteString(" bytes ...\n")
	buf.Write(w.suffix[w.suffixOff:])
	buf.Write(w.suffix[:w.suffixOff])
	return buf.Bytes()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// dedupEnv returns a copy of env with any duplicates removed, in favor of
// later values.
// Items not of the normal environment "key=value" form are preserved unchanged.
func dedupEnv(env []string) []string {
	return dedupEnvCase(runtime.GOOS == "windows", env)
}

// dedupEnvCase is dedupEnv with a case option for testing.
// If caseInsensitive is true, the case of keys is ignored.
func dedupEnvCase(caseInsensitive bool, env []string) []string {
	out := make([]string, 0, len(env))
	saw := map[string]int{} // key => index into out
	for _, kv := range env {
		eq := strings.Index(kv, "=")
		if eq < 0 {
			out = append(out, kv)
			continue
		}
		k := kv[:eq]
		if caseInsensitive {
			k = strings.ToLower(k)
		}
		if dupIdx, isDup := saw[k]; isDup {
			out[dupIdx] = kv
			continue
		}
		saw[k] = len(out)
		out = append(out, kv)
	}
	return out
}
