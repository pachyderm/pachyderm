// Package starlark runs Pachyderm-specific starlark programs.  It's recommended to import this as
// "ourstar" when also using go.starlark.net/starlark in the same package.
package starlark

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	starjson "go.starlark.net/lib/json"
	starmath "go.starlark.net/lib/math"
	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"go.uber.org/zap"
)

const (
	goContextKey       = "goContext"
	scriptDirectoryKey = "scriptDirectory"
)

func init() {
	starlark.Universe["json"] = starjson.Module
	starlark.Universe["time"] = startime.Module
	starlark.Universe["math"] = starmath.Module
}

// GetContext returns a context.Context that applies to in the provided Starlark thread.
func GetContext(t *starlark.Thread) context.Context {
	x := t.Local(goContextKey)
	if ctx, ok := x.(context.Context); ok {
		return ctx
	}
	return pctx.TODO()
}

func scriptDirectory(t *starlark.Thread) string {
	x := t.Local(scriptDirectoryKey)
	if s, ok := x.(string); ok {
		return s
	}
	return "/dev/null"
}

func nameScript(t *starlark.Thread, abs string) string {
	wd, err := os.Getwd()
	if err != nil {
		return abs
	}
	rel, err := filepath.Rel(wd, abs)
	if err != nil {
		return abs
	}
	return rel
}

type loadResult struct {
	globals starlark.StringDict
	err     error
}

func newThread(name string) *starlark.Thread {
	return &starlark.Thread{
		Name: name,
		Print: func(t *starlark.Thread, msg string) {
			log.Info(GetContext(t), msg)
		},
	}
}

// Options controls the Starlark execution environment.
type Options struct {
	// Modules are built-in modules that are available for programs to load with "load".
	// "*.star" is always resolve to a disk path; everything else can be a key in this map.  The
	// StringDict map value should be a map from function names to starlark.Builtin objects.
	Modules map[string]starlark.StringDict
	// Predefined is a dictionary of predefined values that don't need to be loaded.
	Predefined starlark.StringDict
	// REPLPredefined are values that are only predefined in an interactive shell.
	REPLPredefined starlark.StringDict
	// ThreadLocalVars are additional thread-local variables to set.  Go code can get at these
	// environment variables with thread.Local(key).  See GetContext() for an example.
	ThreadLocalVars map[string]any
}

// RunFunc is a function that can run a Starlark program in a thread.  in is the string passed to Run,
// module is the canonical module name for that code, if applicable.
type RunFunc func(fileOpts *syntax.FileOptions, thread *starlark.Thread, in string, module string, globals starlark.StringDict) (starlark.StringDict, error)

// Run runs the provided RunFunc.  The input "in" is not run here, and only serves as a base path
// for resolving module loads.
func Run(rctx context.Context, in string, opts Options, f RunFunc) (starlark.StringDict, error) {
	ctx, c := pctx.WithCancel(rctx)
	defer c()
	path := filepath.Clean(in)
	dir := filepath.Dir(path)
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "find absolute location of starlark file %v", path)
	}

	thread := newThread(filepath.Base(path))
	thread.SetLocal(goContextKey, ctx)
	thread.SetLocal(scriptDirectoryKey, abs)

	fileOpts := &syntax.FileOptions{
		Set:               true,
		While:             true,
		TopLevelControl:   true,
		GlobalReassign:    true,
		LoadBindsGlobally: false,
		Recursion:         false,
	}
	vars := opts.Predefined
	if vars == nil {
		vars = make(starlark.StringDict)
	}
	if opts.Modules == nil {
		opts.Modules = make(map[string]starlark.StringDict)
	}

	modules := make(map[string]*loadResult)
	var load func(*starlark.Thread, string) (starlark.StringDict, error)
	load = func(t *starlark.Thread, module string) (starlark.StringDict, error) {
		// If this is a built-in module, just return it.
		if !strings.HasSuffix(module, ".star") {
			if data, ok := opts.Modules[module]; ok {
				return data, nil
			} else {
				return nil, errors.Errorf("no built-in module %q", module)
			}
		}
		script := filepath.Clean(filepath.Join(scriptDirectory(t), module))
		dir, err := filepath.Abs(filepath.Dir(script))
		if err != nil {
			return nil, errors.Wrapf(err, "setup location of module %v", module)
		}
		module = script

		// Check to see if we've already loaded the module.
		result, ok := modules[module]
		if ok {
			// result=nil is a placeholder indicating that we are already attempting to
			// load this module.
			if result == nil {
				return nil, errors.New("cycle in load graph")
			}
			// Otherwise, load from the cache.
			return result.globals, result.err
		}
		// Mark this module as attempting to load.  This will signal an import cycle.
		modules[module] = nil

		// Since we're here, we want to read a file from disk and interpret it.
		name := nameScript(thread, module)
		ctx, done := log.SpanContext(GetContext(t), "load("+name+")")
		localThread := newThread(name)
		localThread.Load = load
		localThread.SetLocal(goContextKey, ctx)
		localThread.SetLocal(scriptDirectoryKey, dir)
		for k, v := range opts.ThreadLocalVars {
			localThread.SetLocal(k, v)
		}

		log.Debug(ctx, "attempting to load module from disk")
		data, err := os.ReadFile(module)
		if err != nil {
			err = errors.Wrap(err, "read module")
			done(zap.Error(err))
			return nil, err
		}
		threadVars := make(starlark.StringDict)
		for k, v := range vars {
			threadVars[k] = v
		}
		globals, err := starlark.ExecFileOptions(fileOpts, localThread, name, data, threadVars)
		modules[module] = &loadResult{
			globals: globals,
			err:     err,
		}
		err = errors.Wrap(err, "compile module")
		done(Error("error", err)...)
		if err != nil {
			return nil, err
		}
		return globals, nil
	}
	thread.Load = load
	for k, v := range opts.ThreadLocalVars {
		thread.SetLocal(k, v)
	}
	go func() {
		<-ctx.Done()
		if err := context.Cause(ctx); err != nil {
			thread.Cancel(err.Error())
		} else {
			thread.Cancel("no context error, but context is done")
		}

	}()
	return f(fileOpts, thread, in, path, vars)
}

// RunProgram runs an on-disk Starlark script located at path.
func RunProgram(ctx context.Context, path string, opts Options) (starlark.StringDict, error) {
	return Run(ctx, path, opts, func(fileOpts *syntax.FileOptions, thread *starlark.Thread, _ string, module string, globals starlark.StringDict) (starlark.StringDict, error) {
		result, err := starlark.ExecFileOptions(fileOpts, thread, module, nil, globals)
		if err != nil {
			return result, errors.Wrapf(err, "exec starlark file %v", module)
		}
		return result, nil
	})
}

// RunScript runs a script named "name" from the program text in "script".
func RunScript(ctx context.Context, name string, script string, opts Options) (starlark.StringDict, error) {
	return Run(ctx, name, opts, func(fileOpts *syntax.FileOptions, thread *starlark.Thread, in, module string, globals starlark.StringDict) (starlark.StringDict, error) {
		// If the implementation looks confusing, ExecFileOptions is "generic" in that the
		// "src" argument can be a string or []byte with program text.
		result, err := starlark.ExecFileOptions(fileOpts, thread, name, script, globals)
		if err != nil {
			return result, errors.Wrap(err, "exec starlark script")
		}
		return result, nil

	})
}
