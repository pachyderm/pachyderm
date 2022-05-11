package csv

import (
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestRead(t *testing.T) {
	input := `,null,"",,,"""""",,"hello, world",`
	v1, v2, v3, v4 := "null", "", `""`, "hello, world"
	expected := [][]*string{{nil, &v1, &v2, nil, nil, &v3, nil, &v4, nil}}

	r := NewReader(strings.NewReader(input))
	out, err := r.ReadAll()
	require.NoError(t, err)
	require.Equal(t, expected, out)
}
