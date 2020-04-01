package localcache

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

type entry struct {
	key, val string
}

func put(t *testing.T, c *Cache, e *entry) {
	require.NoError(t, c.Put(e.key, bytes.NewReader([]byte(e.val))))
}

func get(t *testing.T, c *Cache, e *entry) {
	r, err := c.Get(e.key)
	require.NoError(t, err)
	output, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, e.val, string(output))
}

func getError(t *testing.T, c *Cache, e *entry) {
	_, err := c.Get(e.key)
	require.YesError(t, err)
}

func getNotEqual(t *testing.T, c *Cache, e *entry) {
	r, err := c.Get(e.key)
	require.NoError(t, err)
	output, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	require.NotEqual(t, e.val, string(output))
}

func del(t *testing.T, c *Cache, key string) {
	err := c.Delete(key)
	require.NoError(t, err)
	_, err = c.Get(key)
	require.YesError(t, err)
}

func TestSimplePutGet(t *testing.T) {
	c, err := NewCache(".")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, c.Clear())
	}()
	for i := 0; i < 2; i++ {
		e := &entry{strconv.Itoa(i), "val"}
		getError(t, c, e)
		put(t, c, e)
		get(t, c, e)
	}
	e := &entry{"0", "newVal"}
	getNotEqual(t, c, e)
	put(t, c, e)
	get(t, c, e)
}

func TestKeys(t *testing.T) {
	c, err := NewCache(".")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, c.Clear())
	}()
	for i := 0; i < 3; i++ {
		put(t, c, &entry{strconv.Itoa((i + 1) % 3), "val"})
	}
	require.Equal(t, []string{"0", "1", "2"}, c.Keys())
}

func TestDelete(t *testing.T) {
	c, err := NewCache(".")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, c.Clear())
	}()
	var entries []*entry
	for i := 0; i < 3; i++ {
		e := &entry{strconv.Itoa(i), "val"}
		put(t, c, e)
		entries = append(entries, e)
	}
	for i, e := range entries {
		del(t, c, e.key)
		// Check that other entries still exist
		for j := i + 1; j < 3; j++ {
			get(t, c, entries[j])
		}
	}
}

func TestClear(t *testing.T) {
	c, err := NewCache(".")
	require.NoError(t, err)

	defer func() {
		require.NoError(t, c.Clear())
	}()
	for i := 0; i < 3; i++ {
		e := &entry{strconv.Itoa(i), "val"}
		put(t, c, e)
	}
	require.NoError(t, c.Clear())
	for i := 0; i < 3; i++ {
		e := &entry{strconv.Itoa(i), "val"}
		getError(t, c, e)
	}
}
