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
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
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
		object := randutil.UniqueString("test-missing-object-")
		requireExists(t, client, object, false)

		err := client.Get(ctx, object, &bytes.Buffer{})
		require.YesError(t, err)
		require.True(t, pacherr.IsNotExist(err))
	})

	t.Run("TestSingleWrite", func(t *testing.T) {
		t.Parallel()
		client := newClient(t)
		doWriteTest(t, client, randutil.UniqueString("test-single-write-"), []byte("foo bar"))
	})

	t.Run("TestSubdirectory", func(t *testing.T) {
		t.Parallel()
		client := newClient(t)
		object := path.Join(randutil.UniqueString("test-subdirectory-"), "object")
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

func TestEmptyWrite(t *testing.T, client Client) {
	doWriteTest(t, client, randutil.UniqueString("test-empty-write-"), []byte{})
}

// TestInterruption
// Interruption is currently not implemented on the Amazon, Microsoft, and Minio clients
//  Amazon client - use *WithContext methods
//  Microsoft client - move to github.com/Azure/azure-storage-blob-go which supports contexts
//  Minio client - upgrade to v7 which supports contexts in all APIs
//  Local client - interruptible file operations are not a thing in the stdlib
func TestInterruption(t *testing.T, client Client) {
	// Make a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	object := randutil.UniqueString("test-interruption-")
	defer requireExists(t, client, object, false)

	err := client.Put(ctx, object, bytes.NewReader(nil))
	require.YesError(t, err)
	require.ErrorIs(t, err, context.Canceled)

	w := &bytes.Buffer{}
	err = client.Get(ctx, object, w)
	require.YesError(t, err)
	require.ErrorIs(t, err, context.Canceled)

	err = client.Delete(ctx, object)
	require.YesError(t, err)
	require.ErrorIs(t, err, context.Canceled)

	err = client.Walk(ctx, object, func(name string) error {
		require.False(t, true)
		return nil
	})
	require.YesError(t, err)
	require.ErrorIs(t, err, context.Canceled)

	_, err = client.Exists(ctx, object)
	require.YesError(t, err)
	require.ErrorIs(t, err, context.Canceled)
}

// TestStorage is a defensive method for checking to make sure that storage is
// properly configured.
func TestStorage(ctx context.Context, c Client) error {
	testObj := "test/" + uuid.NewWithoutDashes()
	if err := func() (retErr error) {
		data := []byte("test")
		return errors.EnsureStack(c.Put(ctx, testObj, bytes.NewReader(data)))
	}(); err != nil {
		return errors.Wrapf(err, "unable to write to object storage")
	}
	if err := func() (retErr error) {
		buf := bytes.NewBuffer(nil)
		return errors.EnsureStack(c.Get(ctx, testObj, buf))
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

func requireExists(t testing.TB, client Client, object string, expected bool) {
	exists, err := client.Exists(context.Background(), object)
	require.NoError(t, err)
	require.Equal(t, expected, exists)
}

func doWriteTest(t testing.TB, client Client, object string, data []byte) {
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
