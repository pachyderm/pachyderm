package debugstar

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"k8s.io/client-go/kubernetes"
)

// XXX: copied right out of server to break import cycle
type DumpFS interface {
	Write(string, func(io.Writer) error) error
}

type starlarkDumpFS struct {
	DumpFS
	frozen bool
}

var _ starlark.Value = (*starlarkDumpFS)(nil)
var _ starlark.Callable = (*starlarkDumpFS)(nil)

func (s *starlarkDumpFS) Freeze()               { s.frozen = true }
func (s *starlarkDumpFS) Hash() (uint32, error) { return 0, errors.New("dumpfs not hashable") }
func (s *starlarkDumpFS) Truth() starlark.Bool  { return true }
func (s *starlarkDumpFS) String() string        { return "<dumpfs>" }
func (s *starlarkDumpFS) Type() string          { return "referencelist" }
func (s *starlarkDumpFS) Name() string          { return "dump" }

// Starlark: dump(name, content)
func (s *starlarkDumpFS) CallInternal(t *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if s.frozen {
		return nil, errors.New("dump is frozen")
	}
	var filename string
	var value string
	if err := starlark.UnpackArgs("dump", args, kwargs, "filename", &filename, "content", &value); err != nil {
		return nil, errors.Wrap(err, "unpack args")
	}
	if err := s.writeBytes(filename, strings.NewReader(value)); err != nil {
		return nil, errors.Wrap(err, "write bytes")
	}
	return starlark.None, nil
}

func (s *starlarkDumpFS) writeBytes(file string, r io.Reader) error {
	if err := s.DumpFS.Write(file, func(w io.Writer) error {
		if _, err := io.Copy(w, r); err != nil {
			return errors.Wrap(err, "copy")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "Write")
	}
	return nil
}

type Env struct {
	Kubernetes kubernetes.Interface
}

func (e *Env) RunStarlark(rctx context.Context, name string, script string, fs DumpFS) (retErr error) {
	ctx, done := log.SpanContext(rctx, fmt.Sprintf("starlark(%v)", name))
	defer done(log.Errorp(&retErr))
	defer func() {
		if err := recover(); err != nil {
			errors.JoinInto(&retErr, errors.Errorf("starlark evaluation panicked: %v", err))
		}
	}()
	k8s := &starlarkstruct.Module{
		Name:    "k8s",
		Members: starlark.StringDict{},
	}
	if _, err := ourstar.RunScript(ctx, name, script, ourstar.Options{
		Predefined: starlark.StringDict{
			"dump": &starlarkDumpFS{DumpFS: fs},
			"k8s":  k8s,
		},
	}); err != nil {
		return errors.Wrap(err, "RunScript")
	}
	return nil
}
