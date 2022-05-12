package csv

import (
	"bytes"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestWrite(t *testing.T) {
	v1, v2, v3 := "null", "", `""`
	expected := `null,"",,""""""` + "\n"

	b := &bytes.Buffer{}
	f := NewWriter(b)
	err := f.WriteAll([][]*string{{&v1, &v2, nil, &v3}})
	require.NoError(t, err)
	out := b.String()
	require.Equal(t, expected, out)
}
