package ancestry

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

var ancestryTests = []struct {
	in        string
	name      string
	ancestors int
}{
	{"foo", "foo", 0},
	{"foo^0", "foo", 0},
	{"foo^-0", "foo", 0},
	{"foo~0", "foo", 0},
	{"foo.0", "foo", 0},
	{"foo~1000000", "foo", 1000000},
	{"foo^1000000", "foo", 1000000},
	{"foo^-1000000", "foo", -1000000},
	{"foo.1000000", "foo", -1000000},
	{"foo.-1000000", "foo", 1000000},
}

func TestAncestry(t *testing.T) {
	for i, test := range ancestryTests {
		name, ancestors, err := Parse(test.in)
		require.NoError(t, err, "ancestryTests[%d]", i)
		require.Equal(t, test.name, name, "ancestryTests[%d]", i)
		require.Equal(t, test.ancestors, ancestors, "ancestryTests[%d]", i)
	}
}

var validNames = []string{
	"foo",
	"foo2",
	"bar",
	"5bar",
	"foo_bar",
	"foo-bar",
	"1-2-3-4-5-6-7-8-9",
	"1_2-3_4-5_6-7_8-9",
}

var invalidNames = []string{
	"foo+",
	"foo.bar",
	"foo^3",
	"^^^^^",
	">.<",
	"(*)",
	"bizz, buzz",
	"    ",
}

func TestValidate(t *testing.T) {
	for i, name := range validNames {
		require.NoError(t, ValidateName(name), "validNames[%d]", i)
	}
	for i, name := range invalidNames {
		require.YesError(t, ValidateName(name), "invalidNames[%d]", i)
		require.NoError(t, ValidateName(SanitizeName(name)), "invalidNames[%d]", i)
	}
}
