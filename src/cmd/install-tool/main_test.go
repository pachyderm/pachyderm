package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestInstall(t *testing.T) {
	ctx := pctx.TestContext(t)
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	src := filepath.Join(tmp, "src")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatalf("create bin dir: %v", err)
	}
	dst := filepath.Join(bin, "new")
	want := []byte("hello\n")
	if err := os.WriteFile(src, want, 0o644); err != nil {
		t.Fatalf("create src data: %v", err)
	}
	if err := install(ctx, dst, src); err != nil {
		t.Fatalf("install: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst (%v): %v", dst, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("installed content (-want +got):\n%s", diff)
	}
}

func TestBadInstall(t *testing.T) {
	ctx := pctx.TestContext(t)
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	dst := filepath.Join("does", "not", "exist")
	want := []byte("hello\n")
	if err := os.WriteFile(src, want, 0o644); err != nil {
		t.Fatalf("create src data: %v", err)
	}
	if err := install(ctx, dst, src); err == nil {
		t.Errorf("somehow installed to %v ok", dst)
	}
	if _, err := os.ReadFile(dst); err == nil {
		t.Errorf("somehow read destination %v ok", dst)
	}
}

func TestFind(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	if err := os.WriteFile(bin, []byte("already exists\n"), 0o755); err != nil {
		t.Fatalf("create already-existing binary: %v", err)
	}

	testData := []struct {
		name, home string
		path, want []string
	}{
		{
			name: "my $PATH",
			home: "/home/user",
			path: []string{
				"/home/user/.cache/bazelisk/downloads/sha256/62d62c699c1eb9f9be6a88030912a54d19fe45ae29329c7e5c53aba787492522/bin",
				"/usr/local/go/bin",
				"/home/user/.nvm/versions/node/v18.16.1/bin",
				"/home/linuxbrew/.linuxbrew/bin",
				"/home/linuxbrew/.linuxbrew/sbin",
				"/home/user/pach/install",
				"/home/user/bin",
				"/home/user/go/bin",
				"/home/user/.fly/bin",
				"/home/user/.cargo/bin",
				"/home/user/.npm-packages/bin",
				"/home/user/.krew/bin",
				"/snap/bin",
				"/home/user/.local/bin",
				"/home/user/.gem/bin",
				"/home/user/.gem/ruby/2.5.0/bin",
				"/home/user/.dotfiles/bin",
				"/usr/local/scripts",
				"/home/user/.nix-profile/bin",
				"/nix/var/nix/profiles/default/bin",
				"/usr/local/bin",
				"/usr/bin",
				"/bin",
				"/usr/local/games",
				"/usr/games",
			},
			want: []string{
				"/home/user/bin",
				"/home/user/go/bin",
				"/home/user/.cache/bazelisk/downloads/sha256/62d62c699c1eb9f9be6a88030912a54d19fe45ae29329c7e5c53aba787492522/bin",
				"/home/user/.nvm/versions/node/v18.16.1/bin",
				"/home/user/pach/install",
				"/home/user/.fly/bin",
				"/home/user/.cargo/bin",
				"/home/user/.npm-packages/bin",
				"/home/user/.krew/bin",
				"/home/user/.local/bin",
				"/home/user/.gem/bin",
				"/home/user/.gem/ruby/2.5.0/bin",
				"/home/user/.dotfiles/bin",
				"/home/user/.nix-profile/bin",
				"/usr/local/go/bin",
				"/home/linuxbrew/.linuxbrew/bin",
				"/home/linuxbrew/.linuxbrew/sbin",
				"/snap/bin",
				"/usr/local/scripts",
				"/nix/var/nix/profiles/default/bin",
				"/usr/local/bin",
				"/usr/bin",
				"/bin",
				"/usr/local/games",
				"/usr/games",
			},
		},
		{
			name: "my $PATH with extra slashes",
			home: "/home/user",
			path: []string{
				"/home/user/.cache/bazelisk/downloads/sha256/62d62c699c1eb9f9be6a88030912a54d19fe45ae29329c7e5c53aba787492522/bin/",
				"/usr/local/go/bin/",
				"/home/user/.nvm/versions/node/v18.16.1/bin/",
				"/home/linuxbrew/.linuxbrew/bin/",
				"/home/linuxbrew/.linuxbrew/sbin/",
				"/home/user/pach/install/",
				"/home/user/bin/",
				"/home/user/go/bin/",
				"/home/user/.fly/bin/",
				"/home/user/.cargo/bin/",
				"/home/user/.npm-packages/bin/",
				"/home/user/.krew/bin/",
				"/snap/bin/",
				"/home/user/.local/bin/",
				"/home/user/.gem/bin/",
				"/home/user/.gem/ruby/2.5.0/bin/",
				"/home/user/.dotfiles/bin/",
				"/usr/local/scripts/",
				"/home/user/.nix-profile/bin/",
				"/nix/var/nix/profiles/default/bin/",
				"/usr/local/bin/",
				"/usr/bin/",
				"/bin/",
				"/usr/local/games/",
				"/usr/games/",
			},
			want: []string{
				"/home/user/bin",
				"/home/user/go/bin",
				"/home/user/.cache/bazelisk/downloads/sha256/62d62c699c1eb9f9be6a88030912a54d19fe45ae29329c7e5c53aba787492522/bin/",
				"/home/user/.nvm/versions/node/v18.16.1/bin/",
				"/home/user/pach/install/",
				"/home/user/.fly/bin/",
				"/home/user/.cargo/bin/",
				"/home/user/.npm-packages/bin/",
				"/home/user/.krew/bin/",
				"/home/user/.local/bin/",
				"/home/user/.gem/bin/",
				"/home/user/.gem/ruby/2.5.0/bin/",
				"/home/user/.dotfiles/bin/",
				"/home/user/.nix-profile/bin/",
				"/usr/local/go/bin/",
				"/home/linuxbrew/.linuxbrew/bin/",
				"/home/linuxbrew/.linuxbrew/sbin/",
				"/snap/bin/",
				"/usr/local/scripts/",
				"/nix/var/nix/profiles/default/bin/",
				"/usr/local/bin/",
				"/usr/bin/",
				"/bin/",
				"/usr/local/games/",
				"/usr/games/",
			},
		},

		{
			name: "basic homedir",
			home: "/home/user",
			path: []string{"/home/user/go/bin", "/home/user/bin", "/usr/local/go/bin", "/home/user/.whatever/bin"},
			want: []string{"/home/user/go/bin", "/home/user/bin", "/home/user/.whatever/bin", "/usr/local/go/bin"},
		},
		{
			name: "basic homedir with already-existing binary",
			home: "/home/user",
			path: []string{"/home/user/go/bin", "/home/user/bin", "/usr/local/go/bin", "/home/user/.whatever/bin", tmp},
			want: []string{tmp, "/home/user/go/bin", "/home/user/bin", "/home/user/.whatever/bin", "/usr/local/go/bin"},
		},
		{
			name: "no homedir",
			home: "",
			path: []string{"/usr/local/bin", "/usr/bin"},
			want: []string{"/usr/local/bin", "/usr/bin"},
		},
	}

	for _, test := range testData {
		os.Setenv("GOPATH", "/home/user/go")
		t.Run(test.name, func(t *testing.T) {
			got := find(test.home, test.path, "bin")
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("paths (-want +got):\n%s", diff)
			}
		})
	}
}
