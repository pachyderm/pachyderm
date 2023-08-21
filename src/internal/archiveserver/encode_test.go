package archiveserver

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestEncodeV1(t *testing.T) {
	testData := []struct {
		name  string
		paths []string
		want  string
	}{
		{
			name: "empty",
			want: "AQ",
		},
		{
			name:  "one file",
			paths: []string{"default/images@master:/"},
			want:  "ASi1L_0EAMEAAGRlZmF1bHQvaW1hZ2VzQG1hc3RlcjovAGmFDFc",
		},
		{
			name: "doc example",
			paths: []string{
				"default/montage@master:/montage.png",
				"default/images@master:/",
			},
			want: "ASi1L_0EAK0BALQCZGVmYXVsdC9pbWFnZXNAbWFzdGVyOi8AbW9udGFnZW1vbnRhZ2UucG5nAAIQBFwMS4wBy2xQ2w",
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got, err := EncodeV1(test.paths)
			if err != nil {
				t.Fatal(err)
			}
			if want := test.want; got != want {
				t.Errorf("EncodeV1():\n  got: %v\n want: %v", got, want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	testData := []struct {
		name        string
		input       string
		wantVersion uint8
		wantLen     int
		wantErr     bool
	}{
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid base64",
			input:   "\x01\x02\x03\x04",
			wantErr: true,
		},
		{
			name:        "version 1",
			input:       "ATEyMzQ1",
			wantVersion: 1,
			wantLen:     5,
		},
		{
			name:        "version 2",
			input:       "AjEyMzQ1",
			wantVersion: 2,
			wantLen:     5,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			r, version, err := Decode(strings.NewReader(test.input))
			if err != nil && !test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err != nil && test.wantErr {
				return
			} else if err == nil && test.wantErr {
				t.Fatal("expected error, but got success")
			}
			if got, want := version, test.wantVersion; got != want {
				t.Errorf("version:\n  got: %v\n want: %v", got, want)
			}
			content, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}
			if got, want := len(content), test.wantLen; got != want {
				t.Errorf("length after version:\n  got: %v\n want: %v", got, want)
			}
		})
	}
}

func FuzzDecode(f *testing.F) {
	f.Add([]byte("ASi1L_0EaHUBAEQCZGVmYXVsdC9pbWFnZXNAbWFzdGVyOi8AbW9udGFnZS5wbmcAAxQEBQPYsGPLbFDb"))
	f.Fuzz(func(t *testing.T, in []byte) {
		r, version, err := Decode(bytes.NewReader(in))
		if err != nil {
			return
		}
		if version != EncodingVersion1 {
			return
		}
		DecodeV1(r) //nolint:errcheck // We only care about not panicking.
	})
}

func TestDecodeV1Path(t *testing.T) {
	testData := []struct {
		name    string
		path    string
		want    *pfs.File
		wantErr bool
	}{
		{
			name:    "empty",
			path:    "",
			wantErr: true,
		},
		{
			name:    "invalid file, missing project",
			path:    "repo@ref:/path",
			wantErr: true,
		},
		{
			name: "valid branch ref",
			path: "project/repo@ref:/path",
			want: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: "repo",
						Project: &pfs.Project{
							Name: "project",
						},
						Type: "user",
					},
					Branch: &pfs.Branch{
						Repo: &pfs.Repo{
							Name: "repo",
							Project: &pfs.Project{
								Name: "project",
							},
							Type: "user",
						},
						Name: "ref",
					},
				},
				Path: "/path",
			},
		},
		{
			name: "valid commit ref",
			path: "project/repo@44444444444444444444444444444444:/path",
			want: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: "repo",
						Project: &pfs.Project{
							Name: "project",
						},
						Type: "user",
					},
					Id: "44444444444444444444444444444444",
				},
				Path: "/path",
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got, err := DecodeV1Path(test.path)
			if err != nil && !test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && test.wantErr {
				t.Error("expected error, but got success")
			} else if err != nil && test.wantErr {
				t.Logf("got expected error: %v", err)
				return
			}
			if diff := cmp.Diff(got, test.want, protocmp.Transform()); diff != "" {
				t.Errorf("decoded path (-got +want):\n%s", diff)
			}
		})
	}
}

func FuzzDecodeV1Path(f *testing.F) {
	f.Add("project/repo@ref:/path/")
	f.Fuzz(func(t *testing.T, in string) {
		DecodeV1Path(in) //nolint:errcheck // We only care about not panicking.
	})
}
