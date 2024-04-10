package cleanup

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmputil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

var ran bool

func thing() func() error {
	return func() error {
		ran = true
		return nil
	}
}

func errs(err string) func() error {
	return func() error {
		return errors.New(err)
	}
}

func emptyThing() func() error {
	return nil
}

func TestCleaner(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := new(Cleaner)
	c.AddCleanup("err1", errs("1"))
	c.AddCleanup("thing", thing())
	c.AddCleanup("emptyThing", emptyThing())
	c.AddCleanup("err2", errs("2"))

	c2 := new(Cleaner)
	c2.AddCleanup("err3", errs("3"))
	c.Subsume(c2)

	err := c.Cleanup(ctx)
	if err == nil {
		t.Fatal("unexpected success")
	}
	got := strings.Split(err.Error(), "\n")
	want := []string{"/clean up err3: 3/", "/clean up err2: 2/", "/clean up err1: 1/"}
	require.NoDiff(t, want, got, []cmp.Option{cmputil.RegexpStrings()})
}
