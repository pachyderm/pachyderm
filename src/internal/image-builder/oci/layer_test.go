package oci

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/klauspost/compress/zstd"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestLayer(t *testing.T) {
	wantFS := fstest.MapFS{
		"root/hello.txt": {
			Mode:    0o644,
			ModTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Data:    []byte("hello, world!\n"),
		},
		"root/another.txt": {
			Mode:    0o644,
			ModTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Data:    []byte("hello again!\n"),
		},
	}
	layer, err := NewLayerFromFS(wantFS)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := layer.Close(); err != nil {
			t.Fatal(err)
		}
	})

	sha, gotFS := readTarZst(t, layer.Underlying)
	if diff := cmp.Diff(wantFS, gotFS); diff != "" {
		t.Errorf("data (-want +got):\n%s", diff)
	}

	wantDescriptor := v1.Descriptor{
		MediaType: "application/vnd.oci.image.layer.v1.tar+zstd",
		Digest:    sha256Hex(sha),
		Size:      142,
	}
	if diff := cmp.Diff(wantDescriptor, layer.Descriptor); diff != "" {
		t.Errorf("descriptor (-want +got):\n%s", diff)
	}
}

func readTarZst(t *testing.T, underlying string) ([]byte, fstest.MapFS) {
	t.Helper()
	result := make(fstest.MapFS)
	infh, err := os.Open(underlying)
	if err != nil {
		t.Fatalf("open underlying tar.zst: %v", err)
	}
	sha := sha256.New()
	tee := io.TeeReader(infh, sha)
	zst, err := zstd.NewReader(tee)
	if err != nil {
		t.Fatalf("create zstd reader: %v", err)
	}
	tr := tar.NewReader(zst)
	for {
		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("next tar header: %v", err)
		}
		file := new(fstest.MapFile)
		result[header.Name] = file
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, tr); err != nil {
			t.Fatalf("copy file data: %v", err)
		}
		file.ModTime = header.ModTime
		file.Mode = fs.FileMode(header.Mode)
		file.Data = buf.Bytes()
	}
	return sha.Sum(nil), result
}
