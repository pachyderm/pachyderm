package archiveserver

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestArchiveRequest(t *testing.T) {
	testData := []struct {
		name      string
		url       string
		wantPaths []string
		wantErr   bool
	}{
		{
			name: "empty URL",
			url:  "https://pachyderm.example.com/archive/AQ.zip", // 0x01 | base64url == AQ
		},
		{
			name:    "wrong path",
			url:     "https://pachyderm.example.com/foo/AAAA.zip",
			wantErr: true,
		},
		{
			name:    "wrong version",
			url:     "https://pachdyerm.example.com/archive/Ag.zip", // 0x02 | base64url == Ag
			wantErr: true,
		},
		{
			name:    "short url",
			url:     "https://pachyderm.example.com/",
			wantErr: true,
		},
		{
			name:    "shorter url",
			url:     "https://pachyderm.example.com",
			wantErr: true,
		},
		{
			name:    "long url",
			url:     "https://pachyderm.example.com/lots/of/stuff/archive/AQ.zip",
			wantErr: true,
		},
		{
			name:    "missing extension",
			url:     "https://pachyderm.example.com/archive/AQ",
			wantErr: true,
		},
		{
			name:    "unsupported extension",
			url:     "https://pachyderm.example.com/archive/AQ.tar",
			wantErr: true,
		},
		{
			name: "doc example",
			url:  "https://pachyderm.example.com/archive/ASi1L_0EaHUBAEQCZGVmYXVsdC9pbWFnZXNAbWFzdGVyOi8AbW9udGFnZS5wbmcAAxQEBQPYsGPLbFDb.zip",
			wantPaths: []string{
				"default/images@master:/",
				"default/montage@master:/montage.png",
			},
		},
		{
			name: "doc example, different compression settings",
			url:  "https://pachyderm.example.com/archive/ASi1L_0EAK0BALQCZGVmYXVsdC9pbWFnZXNAbWFzdGVyOi8AbW9udGFnZW1vbnRhZ2UucG5nAAIQBFwMS4wBy2xQ2w.zip",
			wantPaths: []string{
				"default/images@master:/",
				"default/montage@master:/montage.png",
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			u, err := url.Parse(test.url)
			if err != nil {
				if test.wantErr {
					t.Logf("url.Parse: got expected error: %v", err)
					return
				}
				t.Fatalf("url.Parse: %v", err)
			}
			a, err := ArchiveFromURL(u)
			if err != nil {
				if test.wantErr {
					t.Logf("ArchiveFromURL: got expected error: %v", err)
					return
				}
				t.Fatalf("ArchiveFromURL: %v", err)
			}
			var got []string
			if err := a.ForEachPath(func(path string) error {
				got = append(got, path)
				return nil
			}); err != nil {
				if test.wantErr {
					t.Logf("ForEachPath: got expected error: %v", err)
					return
				}
				t.Fatalf("ForEachPath: %v", err)
			}
			if test.wantErr {
				t.Fatal("all steps completed ok, but wanted an error")
			}
			if diff := cmp.Diff(got, test.wantPaths); diff != "" {
				t.Errorf("extracted paths (-got +want):\n%s", diff)
			}
		})
	}
}

func FuzzArchiveRequest(f *testing.F) {
	f.Add("https://pachyderm.example.com/archive/ASi1L_0EaHUBAEQCZGVmYXVsdC9pbWFnZXNAbWFzdGVyOi8AbW9udGFnZS5wbmcAAxQEBQPYsGPLbFDb.zip")
	f.Add("https://pachyderm.example.com/archive/ASi1L_0EAK0BALQCZGVmYXVsdC9pbWFnZXNAbWFzdGVyOi8AbW9udGFnZW1vbnRhZ2UucG5nAAIQBFwMS4wBy2xQ2w.zip")
	f.Fuzz(func(t *testing.T, in string) {
		u, err := url.Parse(in)
		if err != nil {
			return
		}
		a, err := ArchiveFromURL(u)
		if err != nil {
			return
		}
		a.ForEachPath(func(path string) error { return nil }) //nolint:errcheck // We only care about not panicking.
	})
}
