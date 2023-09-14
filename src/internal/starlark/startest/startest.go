package startest

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
	"golang.org/x/exp/maps"
)

const testingTKey = "testingT"

// T is a subset of testing.TB for Starlark tests.  We don't attempt to implement testing.TB because
// it contains a private field to prevent people from implementing it.  We also have Run.  We are
// playing these games so that the "failing tests" test suite can run against either a testing.T (to
// see that the fail while developing), or a recorder (to see that they continue to "fail" according
// to the spec; i.e. all the Go tests will pass, when the Go tests are running intentionally-failing
// Starlark tests).
type T interface {
	Run(string, func(T)) bool
	Helper()
	Log(...any)
	Logf(string, ...any)
	Error(...any)
	Errorf(string, ...any)
	Fatal(...any)
	Fatalf(string, ...any)
}

type testingTWrapper struct {
	*testing.T
}

var _ T = (*testingTWrapper)(nil)

func (w *testingTWrapper) Run(name string, f func(T)) bool {
	return w.T.Run(name, func(t *testing.T) {
		f(&testingTWrapper{T: t})
	})
}

// testingT is the testing.T API exposed to Starlark tests.
var testingT = starlark.StringDict{
	// print() is preferred over Log().
	"Logf":   makeStarlarkTestFunction("Logf", func(t T, c string, args []any) { call(t.Log, c, args) }),
	"Error":  makeStarlarkTestFunction("Error", func(t T, c string, args []any) { call(t.Error, c, args) }),
	"Errorf": makeStarlarkTestFunction("Errorf", func(t T, c string, args []any) { callf(t.Errorf, c, args) }),
	"Fatal":  makeStarlarkTestFunction("Fatal", func(t T, c string, args []any) { call(t.Fatal, c, args) }),
	"Fatalf": makeStarlarkTestFunction("Fatal", func(t T, c string, args []any) { callf(t.Fatalf, c, args) }),
}

func call(f func(args ...any), caller string, args []any) {
	if len(args) > 0 {
		args[0] = fmt.Sprintf("%v: %v", caller, args[0])
	} else {
		args = append(args, fmt.Sprintf("%v: %v", caller, "empty call"))
	}
	f(args...)
}

func callf(f func(message string, args ...any), caller string, originalArgs []any) {
	msg := "%v: empty call"
	args := []any{caller}
	if len(originalArgs) > 0 {
		if _, ok := originalArgs[0].(string); ok {
			// TODO: if the filename contains a %, we break the format string.
			msg = fmt.Sprintf("%v: %v", "%v", originalArgs[0])
			args = append(args, originalArgs[1:]...)
		} else {
			msg = "%v: invalid format string: "
			args = append(args, originalArgs...)
		}
	}
	f(msg, args...)
}

func caller(th *starlark.Thread) string {
	stack := th.CallStack()
	caller := strings.ReplaceAll(stack.String(), "\n", "/") // emergency backup
	if len(stack) > 1 {
		i := len(stack) - 2
		caller = fmt.Sprintf("%v:%v:%v", stack[i].Pos.Filename(), stack[i].Pos.Line, stack[i].Pos.Col)
	}
	return caller
}

func makeStarlarkTestFunction(name string, f func(t T, caller string, args []any)) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(th *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		t := getTestingT(th)
		caller := caller(th)
		if len(kwargs) != 0 {
			t.Errorf("%v: unexpected kwargs in %v", caller, fn.Name())
			return starlark.None, nil
		}
		var goArgs []any
		for _, arg := range args {
			v := ourstar.FromStarlark(arg)
			goArgs = append(goArgs, v)
		}
		f(t, caller, goArgs)
		return starlark.None, nil
	})
}

func getTestingT(th *starlark.Thread) T {
	return th.Local(testingTKey).(T)
}

func setTestingT(th *starlark.Thread, t T) {
	th.SetLocal(testingTKey, t)
}

func setupPrint(th *starlark.Thread, t T) {
	// Redirect print() to t.Log.
	th.Print = func(th *starlark.Thread, msg string) {
		t.Logf("%v: %v", caller(th), msg)
	}
}

// RunTest runs an on-disk Starlark test.  Tests are scripts that produce functions that start with
// the name "test_", and take a *testing.T object as their first argument.  The tests work like Go
// tests; t.Error fails the test and continues, t.Fatal aborts the test.  Each Starlark test
// function is run as a subtest of a "starlark" test created by this command.  Starlark tests cannot
// call t.Parallel, since we use the same Starlark thread for each test, but there is no reason why
// this stipulation cannot be removed.
func RunTest(t *testing.T, script string, opts ourstar.Options) {
	ctx := pctx.TestContext(t)
	runTest(ctx, &testingTWrapper{T: t}, script, opts)
}

func runTest(ctx context.Context, t T, script string, opts ourstar.Options) {
	t.Helper()
	t.Run("starlark", func(t T) {
		if _, err := ourstar.Run(ctx, script, opts, func(fileOpts *syntax.FileOptions, th *starlark.Thread, in, module string, globals starlark.StringDict) (starlark.StringDict, error) {
			setupPrint(th, t)
			result, err := starlark.ExecFileOptions(fileOpts, th, module, nil, globals)
			if err != nil {
				return nil, errors.Wrap(err, "compile test")
			}

			// We can't easily investigate the top-level identifiers in source code
			// order; too much information from the parser fails to make it here.  So,
			// we look for valid tests in alphabetical order, and then run those tests
			// in source code order.
			identifiers := maps.Keys(result)
			sort.Strings(identifiers)
			var tests []*starlark.Function
			var skipped []string
			for _, k := range identifiers {
				v := result[k]
				if !strings.HasPrefix(k, "test_") {
					skipped = append(skipped, k)
					continue
				}
				if v.Type() != "function" {
					skipped = append(skipped, k)
					continue
				}
				f, ok := v.(*starlark.Function)
				if !ok {
					skipped = append(skipped, v.String())
					continue
				}
				if got, want := f.NumParams(), 1; got != want {
					t.Errorf("%v@%v: unexpected params: got %v, want %v", f.Name(), f.Position().String(), got, want)
					skipped = append(skipped, f.Name())
					continue
				}
				tests = append(tests, f)
			}
			if len(skipped) != 0 {
				t.Errorf("invalid top-level identifiers found in test file: %v", skipped)
			}
			if len(tests) == 0 {
				t.Fatal("no tests to run")
			}
			// Run the tests, in order.
			sort.Slice(tests, func(i, j int) bool {
				a, b := tests[i].Position(), tests[j].Position()
				if a.Line == b.Line {
					return a.Col < b.Col
				}
				return a.Line < b.Line
			})
			for _, f := range tests {
				t.Run(f.Name(), func(t T) {
					setupPrint(th, t)
					setTestingT(th, t)
					args := starlark.Tuple([]starlark.Value{
						&starlarkstruct.Module{Name: "testing.T", Members: testingT},
					})
					if _, err := starlark.Call(th, f, args, nil); err != nil {
						t.Fatalf("%v@%v: cannot run test: %v", f.Name(), f.Position().String(), err)
					}
				})
			}
			return nil, nil
		}); err != nil {
			t.Fatalf("run starlark: %v", err)
		}
	})
}
