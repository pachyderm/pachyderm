package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
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
	require.True(t, bytes.Contains(buf.Bytes(), []byte("did not expect to find")))
	require.True(t, bytes.Contains(buf.Bytes(), []byte("test")))
}
