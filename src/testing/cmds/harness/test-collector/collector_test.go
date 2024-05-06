package main

import (
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"golang.org/x/exp/maps"
)

func TestReadTestLogic(t *testing.T) {
	input := strings.NewReader(
		`
{"Time":"2023-12-05T13:12:46.786154134-08:00","Action":"output","Package":"pkg","Output":"TestThatIsFine\n"}

Some gibberish I don't want
{"Time":"2023-12-05T13:12:46.786154134-08:00","Action":"output","Package":"pkg","Output":"BenchmarkTestToIgnore\n"}
{"Time":"2023-12-05T13:12:46.786154134-08:00","Action":"output","Package":"pkg","Output":"ExampleAPIClient_POST\n"}
{"Time":"2023-12-05T13:12:46.786154134-08:00","Action":"start","Package":"pkg","Output":"TestWithWrongAction\n"}
{"Time":"2023-12-05T13:12:46.786154134-08:00","Action":"pass","Package":"pkg"}
{"Time":"2023-12-05T13:12:46.786154134-08:00","Action":"output","Package":"pkg","Output":"ok pkg\n"}
{"Time":"2023-12-05T13:12:46.786154134-08:00","Action":"start","Package":"pkg","Output":"? TestWithWrongStart\n"}
`,
	)
	tests, err := readTests(input)
	require.NoError(t, err)
	require.Equal(t, 1, len(maps.Keys(tests)))
	testNames, ok := tests["pkg"]
	require.True(t, ok)
	require.Equal(t, 1, len(testNames))
	require.Equal(t, "TestThatIsFine", testNames[0])
}

func TestSetSubtractionEqual(t *testing.T) {
	tagged := map[string][]string{
		"pkg/one": {"TestNameOne", "TestNameTwo"},
		"pkg/two": {"TestNameOne", "TestNameTwo"},
	}
	result := subtractTestSet(tagged, tagged)
	require.Equal(t, 0, len(maps.Keys(result)))
}

func TestSetSubtractionNone(t *testing.T) {
	tagged := map[string][]string{
		"pkg/one": {"TestNameOne", "TestNameTwo"},
		"pkg/two": {"TestNameOne", "TestNameTwo"},
	}
	result := subtractTestSet(tagged, map[string][]string{})
	require.Equal(t, 2, len(maps.Keys(result)))
	require.ElementsEqual(t, []string{"TestNameOne", "TestNameTwo"}, result["pkg/one"])
	require.ElementsEqual(t, []string{"TestNameOne", "TestNameTwo"}, result["pkg/two"])
}

func TestSetSubtractionDisjoint(t *testing.T) {
	tagged := map[string][]string{
		"pkg/one": {"TestNameOne", "TestNameTwo"},
		"pkg/two": {"TestNameOne", "TestNameTwo"},
	}
	untagged := map[string][]string{
		"pkg/one":   {"TestNameThree", "TestNameFour"},
		"pkg/three": {"TestNameThree", "TestNameFour"},
	}
	result := subtractTestSet(tagged, untagged)
	require.Equal(t, 2, len(maps.Keys(result)))
	require.ElementsEqual(t, []string{"TestNameOne", "TestNameTwo"}, result["pkg/one"])
	require.ElementsEqual(t, []string{"TestNameOne", "TestNameTwo"}, result["pkg/two"])
}

func TestSetSubtractionOverlapping(t *testing.T) {
	tagged := map[string][]string{
		"pkg/one":   {"TestNameOne", "TestNameTwo"},
		"pkg/two":   {"TestNameOne", "TestNameTwo"},
		"pkg/three": {"TestNameThree", "TestNameFour"},
	}
	untagged := map[string][]string{
		"pkg/one":   {"TestNameOne"},
		"pkg/three": {"TestNameThree", "TestNameFour"},
	}
	result := subtractTestSet(tagged, untagged)
	require.Equal(t, 2, len(maps.Keys(result)))
	require.ElementsEqual(t, []string{"TestNameTwo"}, result["pkg/one"])
	require.ElementsEqual(t, []string{"TestNameOne", "TestNameTwo"}, result["pkg/two"])
}
