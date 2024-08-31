package cmds

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestFileIndicatorToReader(t *testing.T) {
	var spec = `{
  "pipeline": {
    "name": "first"
  },
  "input": {
    "pfs": {
      "glob": "/*",
      "repo": "input"
    }
  },
  "parallelism_spec": {
    "constant": "1"
  },
  "transform": {
    "cmd": [ "/bin/bash" ],
    "stdin": [
      "cp /pfs/input/* /pfs/out"
    ]
  }
}
`

	t.Run("file", func(t *testing.T) {
		d := t.TempDir()
		p := filepath.Join(d, "test")
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fmt.Fprint(f, spec); err != nil {
			t.Fatal(err)
		}

		if err := testReader(p); err != nil {
			t.Error(err)
		}
	})

	t.Run("stdin", func(t *testing.T) {
		d := t.TempDir()
		p := filepath.Join(d, "test")
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fmt.Fprint(f, spec); err != nil {
			t.Fatal(err)
		}
		ff, err := os.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		oldStdin := os.Stdin
		t.Cleanup(func() {
			os.Stdin = oldStdin
		})
		os.Stdin = ff

		if err := testReader("-"); err != nil {
			t.Error(err)
		}
	})

	t.Run("url", func(t *testing.T) {
		h := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprint(rw, spec)
		})
		srv := httptest.NewServer(h)

		if err := testReader(srv.URL); err != nil {
			t.Fatal(err)
		}
	})
}

func testReader(indicator string) (retErr error) {
	r, err := fileIndicatorToReadCloser(indicator)
	if err != nil {
		return err
	}
	defer errors.Close(&retErr, r, "close reader")

	rr := ppsutil.NewSpecReader(r)

	var i int
	for {
		spec, err := rr.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		var p pps.CreatePipelineRequest
		if err := protojson.Unmarshal([]byte(spec), &p); err != nil {
			return errors.Wrap(err, "could not unmarshal CreatePipelineRequest")
		}
		if expected, got := uint64(1), p.ParallelismSpec.Constant; expected != got {
			return errors.Errorf("parallelism spec constant: expected %d; got %d", expected, got)
		}
		i++
	}
	if expected, got := 1, i; expected != got {
		return errors.Errorf("expected %d objects; got %d", expected, got)
	}
	return nil
}
