package startest

import (
	"flag"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
)

func TestRunTest(t *testing.T) {
	RunTest(t, "testdata/good_test.star", ourstar.Options{})
}

type op struct {
	Name string
	Args string
}

type fakeTestingT struct {
	*testing.T
	Name string
	Ops  []*op
}

func (t *fakeTestingT) Run(name string, f func(t T)) bool {
	sub := &fakeTestingT{T: t.T, Name: t.Name + "/" + name}
	defer func() {
		if err := recover(); err != nil && err != "fakeTestingT.Fatal" {
			panic(err)
		}
		t.Ops = append(t.Ops, sub.Ops...)
	}()
	f(sub)
	return false
}

func (t *fakeTestingT) Log(args ...any) {
	t.Ops = append(t.Ops, &op{Name: t.Name + ".Log", Args: fmt.Sprint(args...)})
}

func (t *fakeTestingT) Error(args ...any) {
	t.Ops = append(t.Ops, &op{Name: t.Name + ".Error", Args: fmt.Sprint(args...)})
}

func (t *fakeTestingT) Fatal(args ...any) {
	t.Ops = append(t.Ops, &op{Name: t.Name + ".Fatal", Args: fmt.Sprint(args...)})
	panic("fakeTestingT.Fatal")
}

func (t *fakeTestingT) Logf(msg string, args ...any) {
	t.Ops = append(t.Ops, &op{Name: t.Name + ".Logf", Args: fmt.Sprintf(msg, args...)})
}

func (t *fakeTestingT) Errorf(msg string, args ...any) {
	t.Ops = append(t.Ops, &op{Name: t.Name + ".Errorf", Args: fmt.Sprintf(msg, args...)})
}

func (t *fakeTestingT) Fatalf(msg string, args ...any) {
	t.Ops = append(t.Ops, &op{Name: t.Name + ".Fatalf", Args: fmt.Sprintf(msg, args...)})
	panic("fakeTestingT.Fatal")
}

var runFailingTests = flag.Bool("run-failing-tests", false, "If true, run the tests that are supposed to fail.")

func TestRunFailingTests(t *testing.T) {
	var ourT T
	var got *fakeTestingT
	if *runFailingTests {
		ourT = &testingTWrapper{T: t}
	} else {
		got = &fakeTestingT{T: t}
		ourT = got
	}
	ctx := pctx.TestContext(t)
	runTest(ctx, ourT, "testdata/failing_test.star", ourstar.Options{})

	if !*runFailingTests {
		got.T = nil
		want := &fakeTestingT{
			Ops: []*op{
				{
					Name: "/starlark.Logf",
					Args: "testdata/failing_test.star:25:6: this is logged at the top level",
				},
				{
					Name: "/starlark.Errorf",
					Args: "test_extra_parameters@testdata/failing_test.star:17:1: unexpected params: got 3, want 1",
				},
				{
					Name: "/starlark.Errorf",
					Args: "test_invalid_parameters@testdata/failing_test.star:14:1: unexpected params: got 0, want 1",
				},
				{
					Name: "/starlark.Errorf",
					Args: "invalid top-level identifiers found in test file: [test_extra_parameters test_invalid_parameters test_variable this_is_not_a_test]",
				},
				{
					Name: "/starlark/test_error.Error",
					Args: "testdata/failing_test.star:2:12: this is an errorthis too",
				},
				{
					Name: "/starlark/test_error.Errorf",
					Args: "testdata/failing_test.star:3:13: this is also an error",
				},
				{
					Name: "/starlark/test_fatal.Logf",
					Args: "testdata/failing_test.star:6:10: start",
				},
				{
					Name: "/starlark/test_fatal.Fatal",
					Args: "testdata/failing_test.star:7:12: this has failedseverely",
				},
				{
					Name: "/starlark/test_unexpected_kwargs.Errorf",
					Args: "testdata/failing_test.star:12:13: unexpected kwargs in Errorf",
				},
				{
					Name: "/starlark/test_bad_calls.Logf",
					Args: "testdata/failing_test.star:28:10: ",
				},
				{
					Name: "/starlark/test_bad_calls.Error",
					Args: "testdata/failing_test.star:29:12: empty call",
				},
				{
					Name: "/starlark/test_bad_calls.Log",
					Args: "testdata/failing_test.star:30:11: empty call",
				},
				{
					Name: "/starlark/test_bad_calls.Errorf",
					Args: "testdata/failing_test.star:31:13: invalid format string: %!(EXTRA int64=42, int64=43)",
				},
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("test log (-want +got):\n%s", diff)
		}
	}
}

func TestRunEmptyTest(t *testing.T) {
	var ourT T
	var got *fakeTestingT
	if *runFailingTests {
		ourT = &testingTWrapper{T: t}
	} else {
		got = &fakeTestingT{T: t}
		ourT = got
	}
	ctx := pctx.TestContext(t)
	runTest(ctx, ourT, "testdata/empty_test.star", ourstar.Options{})

	if !*runFailingTests {
		got.T = nil
		want := &fakeTestingT{
			Ops: []*op{
				{
					Name: "/starlark.Logf",
					Args: "testdata/empty_test.star:1:6: This is an empty test!",
				},
				{
					Name: "/starlark.Fatal",
					Args: "no tests to run",
				},
			},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("test log (-want +got):\n%s", diff)
		}
	}
}
