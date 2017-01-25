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

func TestTags(t *testing.T) {
	c := getPachClient(t)
	_, err := c.PutObject([]byte("foo"), "bar", "buzz")
	require.NoError(t, err)
	value, err := c.GetTag("bar")
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
	value, err = c.GetTag("buzz")
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
}

func TestManyObjects(t *testing.T) {
	c := getPachClient(t)
	var objects []string
	for i := 0; i < 200; i++ {
		object, err := c.PutObject([]byte(string(i)))
		require.NoError(t, err)
		objects = append(objects, object.Hash)
	}
	for i, hash := range objects {
		value, err := c.GetObject(hash)
		require.NoError(t, err)
		require.Equal(t, []byte(string(i)), value)
	}
}

func getPachClient(t testing.TB) *client.APIClient {
	client, err := client.NewFromAddress("0.0.0.0:30650")
	require.NoError(t, err)
	return client
}
