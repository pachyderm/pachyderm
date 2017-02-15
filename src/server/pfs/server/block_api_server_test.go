package server

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
)

func TestPutGet(t *testing.T) {
	c := getPachClient(t)
	object, err := c.PutObject(strings.NewReader("foo"))
	require.NoError(t, err)
	value, err := c.ReadObject(object.Hash)
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
	objectInfo, err := c.InspectObject(object.Hash)
	require.NoError(t, err)
	require.Equal(t, uint64(3), objectInfo.BlockRef.Range.Upper-objectInfo.BlockRef.Range.Lower)
}

func TestTags(t *testing.T) {
	c := getPachClient(t)
	object, err := c.PutObject(strings.NewReader("foo"), "bar", "fizz")
	require.NoError(t, err)
	require.NoError(t, c.TagObject(object.Hash, "buzz"))
	value, err := c.ReadTag("bar")
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
	value, err = c.ReadTag("fizz")
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
	value, err = c.ReadTag("buzz")
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
}

func TestManyObjects(t *testing.T) {
	c := getPachClient(t)
	var objects []string
	for i := 0; i < 25; i++ {
		object, err := c.PutObject(strings.NewReader(string(i)), string(i))
		require.NoError(t, err)
		objects = append(objects, object.Hash)
	}
	require.NoError(t, c.Compact())
	for i, hash := range objects {
		value, err := c.ReadObject(hash)
		require.NoError(t, err)
		require.Equal(t, []byte(string(i)), value)
		value, err = c.ReadTag(string(i))
		require.NoError(t, err)
		require.Equal(t, []byte(string(i)), value)
	}
}

func TestBigObject(t *testing.T) {
	c := getPachClient(t)
	r := workload.NewReader(rand.New(rand.NewSource(time.Now().UnixNano())), 50*1024*1024)
	object, err := c.PutObject(r)
	require.NoError(t, err)
	value, err := c.ReadObject(object.Hash)
	require.NoError(t, err)
	require.Equal(t, 50*1024*1024, len(value))
	value, err = c.ReadObject(object.Hash)
	require.NoError(t, err)
	require.Equal(t, 50*1024*1024, len(value))
}

func getPachClient(t testing.TB) *client.APIClient {
	client, err := client.NewFromAddress("0.0.0.0:30650")
	require.NoError(t, err)
	return client
}
