package integrationtests

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestMicrosoftClient(t *testing.T) {
	t.Parallel()
	id, secret, container := LoadMicrosoftParameters(t)
	obj.TestSuite(t, func(t testing.TB) obj.Client {
		client, err := obj.NewMicrosoftClient(container, id, secret)
		require.NoError(t, err)
		return client
	})
	t.Run("Interruption", func(t *testing.T) {
		t.Skip("TODO: cancel write in the microsoft client")
		client, err := obj.NewMicrosoftClient(container, id, secret)
		require.NoError(t, err)
		obj.TestInterruption(t, client)
	})
}
