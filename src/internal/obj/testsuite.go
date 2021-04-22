package obj

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"io/ioutil"
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// TestSuite runs tests to ensure the object returned by newClient implements Client
// newClient should register cleanup of the returned object using testing.T.Cleanup
// All of the subtests call t.Parallel, but the top level test does not.
func TestSuite(t *testing.T, newClient func(t testing.TB) Client) {
	ctx := context.Background()
	t.Run("TestStorage", func(t *testing.T) {
		t.Parallel()
		c := newClient(t)
		require.NoError(t, TestStorage(ctx, c))
	})

	t.Run("TestMissingObject", func(t *testing.T) {
		t.Parallel()
		client := newClient(t)
		object := tu.UniqueString("test-missing-object-")
		requireExists(t, client, object, false)

		err := client.Get(ctx, object, &bytes.Buffer{})
		require.YesError(t, err)
		require.True(t, pacherr.IsNotExist(err))
	})

	t.Run("TestEmptyWrite", func(t *testing.T) {
		t.Parallel()
		client := newClient(t)
		doWriteTest(t, client, tu.UniqueString("test-empty-write-"), []byte{})
	})

	t.Run("TestSingleWrite", func(t *testing.T) {
		t.Parallel()
		client := newClient(t)
		doWriteTest(t, client, tu.UniqueString("test-single-write-"), []byte("foo bar"))
	})

	t.Run("TestSubdirectory", func(t *testing.T) {
		t.Parallel()
		client := newClient(t)
		object := path.Join(tu.UniqueString("test-subdirectory-"), "object")
		doWriteTest(t, client, object, []byte("foo bar"))
	})

	t.Run("TestIntegrity", func(t *testing.T) {
		t.Parallel()
		client := newClient(t)
		name := "prefix/test-object"
		expectedData, err := ioutil.ReadAll(io.LimitReader(rand.Reader, 1<<20))
		require.NoError(t, err)
		expectedHash := pachhash.Sum(expectedData)
		require.NoError(t, client.Put(ctx, name, bytes.NewReader(expectedData)))
		buf := &bytes.Buffer{}
		require.NoError(t, client.Get(ctx, name, buf))
		actualHash := pachhash.Sum(buf.Bytes())
		require.Equal(t, expectedHash, actualHash)
	})
}

// TestStorage is a defensive method for checking to make sure that storage is
// properly configured.
func TestStorage(ctx context.Context, c Client) error {
	testObj := "test/" + uuid.NewWithoutDashes()
	if err := func() (retErr error) {
		data := []byte("test")
		return c.Put(ctx, testObj, bytes.NewReader(data))
	}(); err != nil {
		return errors.Wrapf(err, "unable to write to object storage")
	}
	if err := func() (retErr error) {
		buf := bytes.NewBuffer(nil)
		return c.Get(ctx, testObj, buf)
	}(); err != nil {
		return errors.Wrapf(err, "unable to read from object storage")
	}
	if err := c.Delete(ctx, testObj); err != nil {
		return errors.Wrapf(err, "unable to delete from object storage")
	}
	// Try reading a non-existent object to make sure our IsNotExist function
	// works.
	buf := bytes.NewBuffer(nil)
	err := c.Get(ctx, uuid.NewWithoutDashes(), buf)
	if !pacherr.IsNotExist(err) {
		return errors.Wrapf(err, "storage is unable to discern NotExist errors, should count as NotExist")
	}
	return nil
}

func requireExists(t *testing.T, client Client, object string, expected bool) {
	exists, err := client.Exists(context.Background(), object)
	require.NoError(t, err)
	require.Equal(t, expected, exists)
}

func doWriteTest(t *testing.T, client Client, object string, data []byte) {
	requireExists(t, client, object, false)
	defer requireExists(t, client, object, false)

	ctx := context.Background()
	err := client.Put(ctx, object, bytes.NewReader(data))
	require.NoError(t, err)

	defer func() {
		require.NoError(t, client.Delete(ctx, object))
	}()

	requireExists(t, client, object, true)

	expected := data
	actualBuf := &bytes.Buffer{}
	err = client.Get(context.Background(), object, actualBuf)
	require.NoError(t, err)
	if len(expected) > 0 {
		require.Equal(t, expected, actualBuf.Bytes())
	}
}
