package remapfs

import (
	"fmt"
	"io/fs"
	"regexp"
	"strconv"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmputil"
	"github.com/pachyderm/pachyderm/v2/src/internal/fsutil"
)

func TestEvalSubst(t *testing.T) {
	testData := []struct {
		rx       *regexp.Regexp
		template string
		inputs   []string
		want     []string
	}{
		{
			rx:       regexp.MustCompile("^go_binary$"),
			template: "app/pachd",
			inputs:   []string{"foo", "bar", "baz", "go_binary", "quux"},
			want:     []string{"app/pachd"},
		},
		{
			rx:       regexp.MustCompile("^some-archive-name/(.*)$"),
			template: "$1",
			inputs: []string{"unexpected", "some-archive-name/foo", "some-archive-name/directory",
				"some-archive-name/directory/bar"},
			want: []string{"foo", "directory", "directory/bar"},
		},
		{
			rx:       regexp.MustCompile("^x/(.*)[.](txt|json)$"),
			template: "$2/$1.$2",
			inputs: []string{"x", "x/foo", "x/foo.txt", "x/foo.txt/bar", "y/foo.json",
				"x/foo.json", "x/directory/foo.txt", "x/deeply/nested/directory/foo.json"},
			want: []string{"txt/foo.txt", "json/foo.json", "txt/directory/foo.txt", "json/deeply/nested/directory/foo.json"},
		},
		{
			rx:       regexp.MustCompile("^(.)(.)(.)$"),
			template: "$1🌐$2🌐$3",
			inputs:   []string{"a", "ab", "abc", "abcd", "🌏🌏🌏"},
			want:     []string{"a🌐b🌐c", "🌏🌐🌏🌐🌏"},
		},
		{
			rx:       regexp.MustCompile("^(.)(.)(.)(.)(.)(.)(.)(.)(.)(.)(.)"),
			template: "$11",
			inputs:   []string{"xxxxxxxxxxA"},
			want:     []string{"A"},
		},

		{
			rx:       regexp.MustCompile("abc(.)"),
			template: "$1",
			inputs:   []string{"abcdabce"},
			want:     []string{"d"},
		},
		{
			rx:       regexp.MustCompile("^(.)"),
			template: "$100foo",
			inputs:   []string{"a", "abc"},
		},
		{
			rx:       regexp.MustCompile("a"),
			template: "$foo",
			inputs:   []string{"a", "aa", "aaa"},
		},
		{
			rx:       regexp.MustCompile("a"),
			template: "$42",
			inputs:   []string{"a", "aa", "aaa"},
		},
	}

	for _, test := range testData {
		t.Run(fmt.Sprintf("%v -> %v", test.rx, test.template), func(t *testing.T) {
			var got []string
			for _, in := range test.inputs {
				x, ok := evalSubst(test.rx, test.template, in)
				if ok {
					got = append(got, x)
				}
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("substitutions (-want +got):\n%s", diff)
			}
		})
	}
}

func FuzzTemplates(f *testing.F) {
	f.Add("$1$2$3")
	f.Fuzz(func(t *testing.T, a string) {
		evalSubst(regexp.MustCompile("(.)(.)(.)"), a, a)
	})
}

func TestChmod(t *testing.T) {
	testData := []struct {
		in, chmod, mask, want fs.FileMode
	}{
		{0, 0o777, 0o777, 0o777},
		{0o777, 0, 0, 0o777},
		{0o644, 0o777, 0o111, 0o755},
		{0o755, 0o777, 0o111, 0o755},
	}

	for i, test := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := chmod(test.in, test.chmod, test.mask)
			if want := test.want; got != want {
				t.Errorf("chmod(%b, %b, %b):\n  got: %b\n want: %b", test.in, test.chmod, test.mask, got, want)
			}
		})
	}
}

func TestRemapFS(t *testing.T) {
	nobody := uint32(65535)
	testData := []struct {
		name    string
		input   fs.FS
		rs      []Recipe
		want    []fs.FileInfo
		wantErr string
	}{
		{
			name: "move a go binary into its in-container location",
			input: fstest.MapFS{
				"go_binary": &fstest.MapFile{
					Mode: 0,
				},
			},
			rs: []Recipe{
				{
					Match:     regexp.MustCompile("^go_binary$"),
					Rename:    "app/pachd",
					Chmod:     0o555,
					ChmodMask: 0o777,
				},
				{
					Chown: &nobody,
					Chgrp: &nobody,
				},
			},
			want: []fs.FileInfo{
				&remapFileInfo{
					name: ".",
					mode: fs.ModeDir,
				},
				&remapFileInfo{
					name: "app",
					own:  &nobody,
					grp:  &nobody,
					mode: fs.ModeDir | 0o777,
					size: 4096,
				},
				&remapFileInfo{
					name: "app/pachd",
					own:  &nobody,
					grp:  &nobody,
					mode: 0o555,
				},
			},
		},
		{
			name: "confusing directory names",
			input: fstest.MapFS{
				"foo": &fstest.MapFile{Data: []byte("123")},
				"bar": &fstest.MapFile{Data: []byte("123456789")},
			},
			rs: []Recipe{
				{
					Match:  regexp.MustCompile("^foo$"),
					Rename: "test/test",
				},
				{
					Match:  regexp.MustCompile("^bar$"),
					Rename: "testtest",
				},
			},
			want: []fs.FileInfo{
				&remapFileInfo{
					name: ".",
					mode: fs.ModeDir,
				},
				&remapFileInfo{
					name: "test",
					mode: fs.ModeDir | 0o777,
					size: 4096,
				},
				&remapFileInfo{
					name: "test/test",
					size: 3,
				},
				&remapFileInfo{
					name: "testtest",
					size: 9,
				},
			},
		},
		{
			name: "remove prefix",
			input: fstest.MapFS{
				"archive/README.txt":          &fstest.MapFile{},
				"archive/src/foo.go":          &fstest.MapFile{},
				"archive/src/bar.go":          &fstest.MapFile{},
				"archive/src/cmd/foo/main.go": &fstest.MapFile{},
			},
			rs: []Recipe{
				{
					Match:  regexp.MustCompile("^archive/(.+)$"),
					Rename: "$1",
				},
			},
			want: []fs.FileInfo{
				&remapFileInfo{name: ".", mode: fs.ModeDir},
				&remapFileInfo{name: "README.txt"},
				&remapFileInfo{name: "src", mode: fs.ModeDir},
				&remapFileInfo{name: "src/bar.go"},
				&remapFileInfo{name: "src/cmd", mode: fs.ModeDir},
				&remapFileInfo{name: "src/cmd/foo", mode: fs.ModeDir},
				&remapFileInfo{name: "src/cmd/foo/main.go"},
				&remapFileInfo{name: "src/foo.go"},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			remap := &RemapFS{
				Recipes: test.rs,
				fs:      test.input,
			}

			// Do a TestFS (from the standard library).
			var wantNames []string
			for _, i := range test.want {
				if i.Name() == "." {
					continue
				}
				wantNames = append(wantNames, i.Name())
			}
			if err := fstest.TestFS(remap, wantNames...); err != nil {
				t.Errorf("TestFS: %v", err)
			}

			// Also do our own test of the actual metadata.
			got, err := fsutil.Find(remap)
			if ok := cmputil.WantErr(t, err, test.wantErr); !ok {
				return
			}
			if diff := cmp.Diff(test.want, got, cmp.Transformer("fileinfo", transformFileInfo)); diff != "" {
				t.Errorf("returned files (-want +got):\n%s", diff)
			}
		})
	}
}

func transformFileInfo(i fs.FileInfo) string {
	return fmt.Sprintf("%s", i)
}
