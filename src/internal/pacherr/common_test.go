package pacherr

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestIsNotExist(t *testing.T) {
	err := NewNotExist("collection", "id")
	require.True(t, IsNotExist(err))
}

func TestIsExist(t *testing.T) {
	err := NewExists("collection", "id")
	require.True(t, IsExists(err))
}
