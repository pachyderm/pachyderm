package testing

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/testpachd"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
)

func TestPutGet(t *testing.T) {
	t.Parallel()
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		object, _, err := env.PachClient.PutObject(strings.NewReader("foo"))
		require.NoError(t, err)
		value, err := env.PachClient.ReadObject(object.Hash)
		require.NoError(t, err)
		require.Equal(t, []byte("foo"), value)
		objectInfo, err := env.PachClient.InspectObject(object.Hash)
		require.NoError(t, err)
		require.Equal(t, uint64(3), objectInfo.BlockRef.Range.Upper-objectInfo.BlockRef.Range.Lower)

		return nil
	})
	require.NoError(t, err)
}

func TestTags(t *testing.T) {
	t.Parallel()
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		object, _, err := env.PachClient.PutObject(strings.NewReader("foo"), "bar", "fizz")
		require.NoError(t, err)
		require.NoError(t, env.PachClient.TagObject(object.Hash, "buzz"))
		_, _, err = env.PachClient.PutObject(strings.NewReader("foo"), "quux")
		require.NoError(t, err)
		value, err := env.PachClient.ReadTag("bar")
		require.NoError(t, err)
		require.Equal(t, []byte("foo"), value)
		value, err = env.PachClient.ReadTag("fizz")
		require.NoError(t, err)
		require.Equal(t, []byte("foo"), value)
		value, err = env.PachClient.ReadTag("buzz")
		require.NoError(t, err)
		require.Equal(t, []byte("foo"), value)
		value, err = env.PachClient.ReadTag("quux")
		require.NoError(t, err)
		require.Equal(t, []byte("foo"), value)

		return nil
	})
	require.NoError(t, err)
}

func TestShortTag(t *testing.T) {
	t.Parallel()
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		object, _, err := env.PachClient.PutObject(strings.NewReader("content"), "t")
		require.NoError(t, err)

		value, err := env.PachClient.ReadObject(object.Hash)
		require.NoError(t, err)
		require.Equal(t, []byte("content"), value)
		value, err = env.PachClient.ReadTag("t")
		require.NoError(t, err)
		require.Equal(t, []byte("content"), value)

		return nil
	})
	require.NoError(t, err)
}

func TestManyObjects(t *testing.T) {
	t.Parallel()
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		var objects []string
		for i := 0; i < 25; i++ {
			object, _, err := env.PachClient.PutObject(strings.NewReader(fmt.Sprint(i)), fmt.Sprint(i))
			require.NoError(t, err)
			objects = append(objects, object.Hash)
		}
		for i, hash := range objects {
			value, err := env.PachClient.ReadObject(hash)
			require.NoError(t, err)
			require.Equal(t, []byte(fmt.Sprint(i)), value)
			value, err = env.PachClient.ReadTag(fmt.Sprint(i))
			require.NoError(t, err)
			require.Equal(t, []byte(fmt.Sprint(i)), value)
		}

		return nil
	})
	require.NoError(t, err)
}

func TestBigObject(t *testing.T) {
	t.Parallel()
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		r := workload.NewReader(rand.New(rand.NewSource(time.Now().UnixNano())), 50*1024*1024)
		object, _, err := env.PachClient.PutObject(r)
		require.NoError(t, err)
		value, err := env.PachClient.ReadObject(object.Hash)
		require.NoError(t, err)
		require.Equal(t, 50*1024*1024, len(value))
		value, err = env.PachClient.ReadObject(object.Hash)
		require.NoError(t, err)
		require.Equal(t, 50*1024*1024, len(value))

		return nil
	})
	require.NoError(t, err)
}

func TestReadObjects(t *testing.T) {
	t.Parallel()
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		fooObject, _, err := env.PachClient.PutObject(strings.NewReader("foo"))
		require.NoError(t, err)
		barObject, _, err := env.PachClient.PutObject(strings.NewReader("bar"))
		require.NoError(t, err)
		value, err := env.PachClient.ReadObjects([]string{fooObject.Hash, barObject.Hash}, 0, 0)
		require.NoError(t, err)
		require.Equal(t, []byte("foobar"), value)
		value, err = env.PachClient.ReadObjects([]string{fooObject.Hash, barObject.Hash}, 0, 2)
		require.NoError(t, err)
		require.Equal(t, []byte("fo"), value)
		value, err = env.PachClient.ReadObjects([]string{fooObject.Hash, barObject.Hash}, 0, 4)
		require.NoError(t, err)
		require.Equal(t, []byte("foob"), value)
		value, err = env.PachClient.ReadObjects([]string{fooObject.Hash, barObject.Hash}, 2, 0)
		require.NoError(t, err)
		require.Equal(t, []byte("obar"), value)
		value, err = env.PachClient.ReadObjects([]string{fooObject.Hash, barObject.Hash}, 4, 0)
		require.NoError(t, err)
		require.Equal(t, []byte("ar"), value)

		return nil
	})
	require.NoError(t, err)
}

func TestDirectObjects(t *testing.T) {
	t.Parallel()
	err := testpachd.WithRealEnv(func(env *testpachd.RealEnv) error {
		fmt.Printf("test path: %s\n", env.Directory)
		path := "direct-foo"
		data := []byte("foo")
		fooWriter, err := env.PachClient.DirectObjWriter(path)
		_, err = fooWriter.Write(data)
		require.NoError(t, err)
		require.NoError(t, fooWriter.Close())

		fooReader, err := env.PachClient.DirectObjReader(path)
		readData := &bytes.Buffer{}
		_, err = readData.ReadFrom(fooReader)
		require.NoError(t, err)
		require.Equal(t, data, readData.Bytes())
		require.NoError(t, fooReader.Close())

		return nil
	})
	require.NoError(t, err)
}
