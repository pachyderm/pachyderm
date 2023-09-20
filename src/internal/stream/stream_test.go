package stream

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestIsEOS(t *testing.T) {
	require.True(t, IsEOS(EOS()))
	require.False(t, errors.Is(EOS(), EOS()))
}
