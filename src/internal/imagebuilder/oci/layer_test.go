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
	"github.com/opencontainers/go-digest"
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

	csha, ucsha, gotFS := readTarZst(t, layer.Underlying)
	if diff := cmp.Diff(wantFS, gotFS); diff != "" {
		t.Errorf("data (-want +got):\n%s", diff)
	}

	want := &Layer{
		Underlying: layer.Underlying,
		SHA256:     [32]byte(csha),
		DiffID:     [32]byte(ucsha),
		Descriptor: v1.Descriptor{
			MediaType: "application/vnd.oci.image.layer.v1.tar+zstd",
			Digest:    digest.NewDigestFromBytes(digest.SHA256, csha),
			Size:      158,
		},
	}
	if diff := cmp.Diff(want, layer); diff != "" {
		t.Errorf("layer (-want +got):\n%s", diff)
	}
}

func readTarZst(t *testing.T, underlying string) ([]byte, []byte, fstest.MapFS) {
	t.Helper()
	result := make(fstest.MapFS)
	infh, err := os.Open(underlying)
	if err != nil {
		t.Fatalf("open underlying tar.zst: %v", err)
	}
	csha := sha256.New()
	ctee := io.TeeReader(infh, csha)
	zst, err := zstd.NewReader(ctee)
	if err != nil {
		t.Fatalf("create zstd reader: %v", err)
	}
	ucsha := sha256.New()
	uctee := io.TeeReader(zst, ucsha)
	tr := tar.NewReader(uctee)
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
	return csha.Sum(nil), ucsha.Sum(nil), result
}
