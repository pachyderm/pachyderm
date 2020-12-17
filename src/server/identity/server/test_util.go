package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

// DeleteAll deletes all data in the cluster, including the identity server.
// TODO: DeleteAll should be part of the client DeleteAll method,
// but it fails if postgres isn't running. Once we run postgres for
// all tests, remove this.
func DeleteAll(t testing.TB) {
	client := tu.GetAuthenticatedPachClient(t, tu.AdminUser)
	_, err := client.IdentityAPIClient.DeleteAll(client.Ctx(), &identity.DeleteAllRequest{})
	require.NoError(t, err)
	tu.DeleteAll(t)
}
