package main

import (
	"bytes"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutilpachctl"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestMatchBasic(t *testing.T) {
	require.NoError(t, tu.BashCmd(`
    echo "This is a test" \
      | match "test" \
			| match "This is a test"
	  `).Run())
}

func TestMatchInvert(t *testing.T) {
	c := tu.BashCmd(`
		echo "This is a test" \
		  | match -v "blag" \
			| match "This is a test"
		`)
	c.Stdout = os.Stdout
	require.NoError(t, c.Run())
}

func TestMatchFail(t *testing.T) {
	c := tu.BashCmd(`
		echo "This is a test" \
		  | match "blag" \
			| match "This is a test"
		`)
	buf := &bytes.Buffer{}
	c.Stderr = buf
	require.YesError(t, c.Run())
	require.True(t, bytes.Contains(buf.Bytes(), []byte("failed to find")))
	require.True(t, bytes.Contains(buf.Bytes(), []byte("blag")))
}

func TestMatchInvertedFail(t *testing.T) {
	c := tu.BashCmd(`
		echo "This is a test" \
		  | match -v "test" \
			| match "This is a test"
		`)
	buf := &bytes.Buffer{}
	c.Stderr = buf
	require.YesError(t, c.Run())
	require.True(t, bytes.Contains(buf.Bytes(), []byte("failed to find")))
	require.True(t, bytes.Contains(buf.Bytes(), []byte("test")))
}
