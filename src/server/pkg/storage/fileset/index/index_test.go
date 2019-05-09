package index

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

func Write(tb testing.TB, chunks *chunk.Storage, rangeSize int64, fileNames []string) io.Reader {
	iw := NewWriter(context.Background(), chunks, rangeSize)
	for _, fileName := range fileNames {
		hdr := &Header{
			Hdr: &tar.Header{Name: fileName},
		}
		require.NoError(tb, iw.WriteHeader(hdr))
	}
	r, err := iw.Close()
	require.NoError(tb, err)
	return r
}

func Actual(tb testing.TB, chunks *chunk.Storage, r io.Reader, prefix string) []string {
	ir := NewReader(context.Background(), chunks, r, prefix)
	result := []string{}
	for {
		hdr, err := ir.Next()
		if err == io.EOF {
			return result
		}
		require.NoError(tb, err)
		result = append(result, hdr.Hdr.Name)
	}
}

func Expected(fileNames []string, prefix string) []string {
	result := []string{}
	for _, fileName := range fileNames {
		if strings.HasPrefix(fileName, prefix) {
			result = append(result, fileName)
		}
	}
	return result
}

func Check(t *testing.T, permString string) {
	objC, chunks := chunk.LocalStorage(t)
	defer func() {
		chunks.DeleteAll(context.Background())
		objC.Delete(context.Background(), chunk.Prefix)
	}()
	fileNames := Generate(permString)
	buf := &bytes.Buffer{}
	io.Copy(buf, Write(t, chunks, 10000, fileNames))
	r := bytes.NewReader(buf.Bytes())
	t.Run("Full", func(t *testing.T) {
		prefix := ""
		expected := fileNames
		actual := Actual(t, chunks, r, prefix)
		require.Equal(t, expected, actual)
	})
	t.Run("FirstFile", func(t *testing.T) {
		r.Seek(0, io.SeekStart)
		prefix := fileNames[0]
		expected := []string{prefix}
		actual := Actual(t, chunks, r, prefix)
		require.Equal(t, expected, actual)
	})
	t.Run("FirstRange", func(t *testing.T) {
		r.Seek(0, io.SeekStart)
		prefix := string(fileNames[0][0])
		expected := Expected(fileNames, prefix)
		actual := Actual(t, chunks, r, prefix)
		require.Equal(t, expected, actual)
	})
	t.Run("MiddleFile", func(t *testing.T) {
		r.Seek(0, io.SeekStart)
		prefix := fileNames[len(fileNames)/2]
		expected := []string{prefix}
		actual := Actual(t, chunks, r, prefix)
		require.Equal(t, expected, actual)
	})
	t.Run("MiddleRange", func(t *testing.T) {
		r.Seek(0, io.SeekStart)
		prefix := string(fileNames[len(fileNames)/2][0])
		expected := Expected(fileNames, prefix)
		actual := Actual(t, chunks, r, prefix)
		require.Equal(t, expected, actual)
	})
	t.Run("LastFile", func(t *testing.T) {
		r.Seek(0, io.SeekStart)
		prefix := fileNames[len(fileNames)-1]
		expected := []string{prefix}
		actual := Actual(t, chunks, r, prefix)
		require.Equal(t, expected, actual)
	})
	t.Run("LastRange", func(t *testing.T) {
		r.Seek(0, io.SeekStart)
		prefix := string(fileNames[len(fileNames)-1][0])
		expected := Expected(fileNames, prefix)
		actual := Actual(t, chunks, r, prefix)
		require.Equal(t, expected, actual)
	})
}

func TestSingleLevel(t *testing.T) {
	Check(t, "abc")
}

func TestMultiLevel(t *testing.T) {
	Check(t, "abcdefg")
}
