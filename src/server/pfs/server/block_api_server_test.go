package server

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
)

func TestPutGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := GetPachClient(t, GetBasicConfig())
	object, _, err := c.PutObject(strings.NewReader("foo"))
	require.NoError(t, err)
	value, err := c.ReadObject(object.Hash)
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
	objectInfo, err := c.InspectObject(object.Hash)
	require.NoError(t, err)
	require.Equal(t, uint64(3), objectInfo.BlockRef.Range.Upper-objectInfo.BlockRef.Range.Lower)
}

func TestTags(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := GetPachClient(t, GetBasicConfig())
	object, _, err := c.PutObject(strings.NewReader("foo"), "bar", "fizz")
	require.NoError(t, err)
	require.NoError(t, c.TagObject(object.Hash, "buzz"))
	_, _, err = c.PutObject(strings.NewReader("foo"), "quux")
	require.NoError(t, err)
	value, err := c.ReadTag("bar")
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
	value, err = c.ReadTag("fizz")
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
	value, err = c.ReadTag("buzz")
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
	value, err = c.ReadTag("quux")
	require.NoError(t, err)
	require.Equal(t, []byte("foo"), value)
}

func TestShortTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := GetPachClient(t, GetBasicConfig())
	object, _, err := c.PutObject(strings.NewReader("content"), "t")
	require.NoError(t, err)

	value, err := c.ReadObject(object.Hash)
	require.NoError(t, err)
	require.Equal(t, []byte("content"), value)
	value, err = c.ReadTag("t")
	require.NoError(t, err)
	require.Equal(t, []byte("content"), value)
}

func TestManyObjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := GetPachClient(t, GetBasicConfig())
	var objects []string
	for i := 0; i < 25; i++ {
		object, _, err := c.PutObject(strings.NewReader(string(i)), fmt.Sprint(i))
		require.NoError(t, err)
		objects = append(objects, object.Hash)
	}
	for i, hash := range objects {
		value, err := c.ReadObject(hash)
		require.NoError(t, err)
		require.Equal(t, []byte(string(i)), value)
		value, err = c.ReadTag(fmt.Sprint(i))
		require.NoError(t, err)
		require.Equal(t, []byte(string(i)), value)
	}
}

func TestBigObject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := GetPachClient(t, GetBasicConfig())
	r := workload.NewReader(rand.New(rand.NewSource(time.Now().UnixNano())), 50*1024*1024)
	object, _, err := c.PutObject(r)
	require.NoError(t, err)
	value, err := c.ReadObject(object.Hash)
	require.NoError(t, err)
	require.Equal(t, 50*1024*1024, len(value))
	value, err = c.ReadObject(object.Hash)
	require.NoError(t, err)
	require.Equal(t, 50*1024*1024, len(value))
}

func TestReadObjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c := GetPachClient(t, GetBasicConfig())
	fooObject, _, err := c.PutObject(strings.NewReader("foo"))
	require.NoError(t, err)
	barObject, _, err := c.PutObject(strings.NewReader("bar"))
	require.NoError(t, err)
	value, err := c.ReadObjects([]string{fooObject.Hash, barObject.Hash}, 0, 0)
	require.NoError(t, err)
	require.Equal(t, []byte("foobar"), value)
	value, err = c.ReadObjects([]string{fooObject.Hash, barObject.Hash}, 0, 2)
	require.NoError(t, err)
	require.Equal(t, []byte("fo"), value)
	value, err = c.ReadObjects([]string{fooObject.Hash, barObject.Hash}, 0, 4)
	require.NoError(t, err)
	require.Equal(t, []byte("foob"), value)
	value, err = c.ReadObjects([]string{fooObject.Hash, barObject.Hash}, 2, 0)
	require.NoError(t, err)
	require.Equal(t, []byte("obar"), value)
	value, err = c.ReadObjects([]string{fooObject.Hash, barObject.Hash}, 4, 0)
	require.NoError(t, err)
	require.Equal(t, []byte("ar"), value)
}
