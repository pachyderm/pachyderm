package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestPutGet(t *testing.T) {
	c := getPachClient(t)
	object, err := c.PutObject([]byte("foo"))
	require.NoError(t, err)
	value, err := c.GetObject(object.Hash)
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
}

func getPachClient(t testing.TB) *client.APIClient {
	client, err := client.NewFromAddress("0.0.0.0:30650")
	require.NoError(t, err)
	return client
}
