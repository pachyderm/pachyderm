package csv

import (
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestRead(t *testing.T) {
	input := `,null,"",,,"""""",,旰綷罨袢捺悲ȅ-@烱ęssĂZ稈V`
	v1, v2, v3, v4 := "null", "", `""`, "旰綷罨袢捺悲ȅ-@烱ęssĂZ稈V"
	expected := [][]*string{{nil, &v1, &v2, nil, nil, &v3, nil, &v4}}

	r := NewReader(strings.NewReader(input))
	out, err := r.ReadAll()
	require.NoError(t, err)
	require.Equal(t, expected, out)
}
