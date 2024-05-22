package testutil

import (
	"net"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// Listen creates a new net.Listener on localhost and ensures the listener is closed when the test exits
func Listen(t testing.TB) net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { l.Close() })
	return l
}
