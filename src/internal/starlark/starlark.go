// Package starlark runs Pachyderm-specific starlark programs.
package starlark

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"go.uber.org/zap"
)

const (
	goContextKey       = "goContext"
	scriptDirectoryKey = "scriptDirectory"
)

// Modules that are available in all starpach programs.  The convention here is that built-in
// modules are named "module", while modules in files are named something like "module.star".
var Modules = map[string]starlark.StringDict{}

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

type Options struct {
	Predefined      starlark.StringDict
	REPLPredefined  starlark.StringDict
	ThreadLocalVars map[string]any
}

type runFn func(fileOpts *syntax.FileOptions, thread *starlark.Thread, module string, globals starlark.StringDict) (starlark.StringDict, error)

func run(rctx context.Context, path string, opts Options, f runFn) (starlark.StringDict, error) {
	ctx, c := pctx.WithCancel(rctx)
	defer c()
	path = filepath.Clean(path)
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
		GlobalReassign:    false,
		LoadBindsGlobally: false,
		Recursion:         false,
	}
	vars := opts.Predefined
	if vars == nil {
		vars = make(starlark.StringDict)
	}

	modules := make(map[string]*loadResult)
	var load func(*starlark.Thread, string) (starlark.StringDict, error)
	load = func(t *starlark.Thread, module string) (starlark.StringDict, error) {
		// If this is a built-in module, just return it.
		if !strings.HasSuffix(module, ".star") {
			if data, ok := Modules[module]; ok {
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
			// result=nil is a place holder indicating that we are already attempting to
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
	return f(fileOpts, thread, path, vars)
}

func RunProgram(ctx context.Context, path string, opts Options) (starlark.StringDict, error) {
	return run(ctx, path, opts, func(fileOpts *syntax.FileOptions, thread *starlark.Thread, module string, globals starlark.StringDict) (starlark.StringDict, error) {
		return starlark.ExecFileOptions(fileOpts, thread, module, nil, globals)
	})
}
