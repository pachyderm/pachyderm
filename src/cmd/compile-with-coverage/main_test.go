package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestResolve(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "src", "a", "b", "c"), 0o755); err != nil {
		t.Fatalf("make source: %v", err)
	}
	content := []byte("hello, world")
	if err := os.WriteFile(filepath.Join(tmp, "src", "a", "a.txt"), content, 0o644); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}
	if err := os.Symlink(filepath.Join(tmp, "src", "a", "a.txt"), filepath.Join(tmp, "src", "a", "b", "c", "c.txt")); err != nil {
		t.Fatalf("link src/a/a.txt -> src/a/b/c/c.txt: %v", err)
	}
	if err := resolve(filepath.Join(tmp, "src"), filepath.Join(tmp, "dst")); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	dst := filepath.Join(tmp, "dst", tmp, "src") // destination is the det path + exact path passed to src
	info, err := os.Lstat(filepath.Join(dst, "a", "b", "c", "c.txt"))
	if err != nil {
		t.Fatalf("get info for dst/a/b/c/c.txt: %v", err)
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		t.Errorf("dst/a/b/c/c.txt is a symlink: %v", info)
	}
	text, err := os.ReadFile(filepath.Join(dst, "a", "b", "c", "c.txt"))
	if err != nil {
		t.Fatalf("read dst/a/b/c/c.txt: %v", err)
	}
	if got, want := text, content; !bytes.Equal(got, want) {
		t.Errorf("c.txt:\n  got: %s\n want: %s", got, want)
	}
}
