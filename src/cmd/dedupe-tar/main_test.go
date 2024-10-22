package main

import (
	"archive/tar"
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/klauspost/compress/zstd"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

var testdata = "abcdefghijklmnopqrstuvwxyz"

func TestRun(t *testing.T) {
	// Build an input archive.
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	files := []*tar.Header{
		{
			Name:     "./etc/",
			Typeflag: tar.TypeDir,
			Uid:      0,
			Gid:      0,
			Uname:    "foo",
			Gname:    "foo",
		},
		{
			Name:  "./etc/foo",
			Uid:   0,
			Gid:   0,
			Uname: "foo",
			Gname: "foo",
			Size:  10,
		},
		{
			Name:  "./etc/foo",
			Uid:   65532,
			Gid:   65532,
			Uname: "nonroot",
			Gname: "nonroot",
			Size:  8,
		},
		{
			Name:     "/var/",
			Typeflag: tar.TypeDir,
			Uid:      1,
			Gid:      1,
		},
		{
			Name:  "/etc/bar",
			Uid:   0,
			Gid:   0,
			Uname: "foo",
			Gname: "foo",
			Size:  1,
		},
		{
			Name:     "/bar/",
			Typeflag: tar.TypeDir,
			Uid:      0,
			Gid:      0,
			Size:     0,
			ModTime:  time.Unix(1234, 0),
		},
		{
			Name:     "bar",
			Typeflag: tar.TypeReg,
			Uid:      0,
			Gid:      0,
			Size:     20,
		},
	}
	for _, h := range files {
		h.Format = tar.FormatUSTAR
		if err := tw.WriteHeader(h); err != nil {
			t.Fatalf("write header %v: %v", h.Name, err)
		}
		if h.Size > 0 {
			n, err := io.Copy(tw, strings.NewReader(testdata[:h.Size]))
			if err != nil {
				t.Fatalf("copy data into %v: %v", h.Name, err)
			}
			if got, want := n, h.Size; got != want {
				t.Errorf("copy data into %v: byte count:\n  got: %v\n want: %v", h.Name, got, want)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}

	// Run the copy.
	result := new(bytes.Buffer)
	if err := run(buf, result); err != nil {
		t.Fatalf("run: %v", err)
	}

	// Check the output content.
	var got []tar.Header
	zr, err := zstd.NewReader(result)
	if err != nil {
		t.Fatalf("new zstd reader: %v", err)
	}
	tr := tar.NewReader(zr)
	for {
		h, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("tr.Next: %v", err)
		}
		got = append(got, *h)
		if h.Size > 0 {
			x, err := io.ReadAll(tr)
			if err != nil {
				t.Fatalf("read %v: %v", h.Name, err)
			}
			if got, want := string(x), testdata[:h.Size]; got != want {
				t.Errorf("content of %v:\n  got: %v\n want: %v", h.Name, got, want)
			}
		}
	}
	want := []tar.Header{
		{
			Typeflag: tar.TypeReg,
			Name:     "etc/foo",
			Size:     10,
			ModTime:  time.Unix(0, 0),
			Format:   tar.FormatUSTAR,
		},
		{
			Typeflag: tar.TypeReg,
			Name:     "etc/bar",
			Size:     1,
			ModTime:  time.Unix(0, 0),
			Format:   tar.FormatUSTAR,
		},
		{
			Typeflag: tar.TypeDir,
			Name:     "bar/",
			ModTime:  time.Unix(1234, 0),
			Format:   tar.FormatUSTAR,
		},
		{
			Typeflag: tar.TypeReg,
			Name:     "bar",
			Size:     20,
			ModTime:  time.Unix(0, 0),
			Format:   tar.FormatUSTAR,
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("diff (-want +got):\n%s", diff)
	}
}
