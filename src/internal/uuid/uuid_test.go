package uuid

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestIsUUIDWithoutDashes(t *testing.T) {
	require.Equal(t, 32, len(NewWithoutDashes()))
	require.True(t, IsUUIDWithoutDashes("09abcd098faa4fd98643023485739adb"))

	// 13 character is 4
	require.False(t, IsUUIDWithoutDashes("09abcd098faaefd98643023485739adb"))

	// Length 32
	require.False(t, IsUUIDWithoutDashes("09abcd098faaefd98643023485739adbabc"))

	// Hexadecimal
	require.False(t, IsUUIDWithoutDashes("09abcd098faa4fd98643023485739xyz"))

	// Generated
	require.True(t, IsUUIDWithoutDashes(NewWithoutDashes()))
}
