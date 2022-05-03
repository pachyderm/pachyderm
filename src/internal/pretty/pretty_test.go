package pretty_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"
)

func TestCommafy(t *testing.T) {
	if expected, actual := "foo", pretty.Commafy("foo"); expected != actual {
		t.Errorf("expected %q; got %q", expected, actual)
	}
	if expected, actual := "", pretty.Commafy([]string{}); expected != actual {
		t.Errorf("expected %q; got %q", expected, actual)
	}
	if expected, actual := "foo", pretty.Commafy([]string{"foo"}); expected != actual {
		t.Errorf("expected %q; got %q", expected, actual)
	}
	if expected, actual := "foo and bar", pretty.Commafy([]string{"foo", "bar"}); expected != actual {
		t.Errorf("expected %q; got %q", expected, actual)
	}
	if expected, actual := "foo, bar and baz", pretty.Commafy([]string{"foo", "bar", "baz"}); expected != actual {
		t.Errorf("expected %q; got %q", expected, actual)
	}
}
