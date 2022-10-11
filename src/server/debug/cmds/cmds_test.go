package cmds

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func TestWithFile(t *testing.T) {
	testData := []struct {
		name     string
		input    []byte
		wantLogs *regexp.Regexp
	}{
		{
			name:     "empty",
			wantLogs: regexp.MustCompile("problem creating gzip reader: EOF"),
		},
		{
			name:     "not_gzipped",
			input:    []byte("this is an ordinary file\n"),
			wantLogs: regexp.MustCompile("problem creating gzip reader"),
		},
		{
			name: "not_tarred",
			input: func() []byte {
				buf := new(bytes.Buffer)
				gw := gzip.NewWriter(buf)
				if _, err := fmt.Fprintf(gw, "this text will be compressed\n"); err != nil {
					panic(err)
				}
				if err := gw.Close(); err != nil {
					panic(err)
				}
				return buf.Bytes()
			}(),
			wantLogs: regexp.MustCompile("problem reading tar header"),
		},
		{
			name: "tgz",
			input: func() []byte {
				buf := new(bytes.Buffer)
				gw := gzip.NewWriter(buf)
				tw := tar.NewWriter(gw)
				for _, file := range []string{"foo", "bar", "baz", "quux"} {
					if err := tw.WriteHeader(&tar.Header{
						Name: file,
						Size: int64(100000 * (len(file) + 1)),
					}); err != nil {
						panic(err)
					}
					for i := 0; i < 100000; i++ {
						if _, err := fmt.Fprintf(tw, "%s\n", file); err != nil {
							panic(err)
						}
					}
				}
				if err := tw.Close(); err != nil {
					panic(err)
				}
				if err := gw.Close(); err != nil {
					panic(err)
				}
				return buf.Bytes()
			}(),
			wantLogs: regexp.MustCompile("(?s)receiving sub-file foo.*done.*receiving sub-file bar.*done"),
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			name := filepath.Join(t.TempDir(), test.name)
			logs := new(bytes.Buffer)
			err := withFile(logs, name, func(w io.Writer) error {
				_, err := io.Copy(w, bytes.NewReader(test.input))
				return errors.EnsureStack(err)
			})
			if err != nil {
				t.Fatal(err)
			}

			got, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, test.input, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("downloaded file content:\n%s", diff)
			}

			if test.wantLogs != nil {
				if !test.wantLogs.MatchString(logs.String()) {
					t.Error("logs did not match regexp")
					for _, l := range strings.Split(logs.String(), "\n") {
						t.Log(l)
					}
				}
			}
		})
	}
}
