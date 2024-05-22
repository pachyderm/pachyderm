package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
)

func TestQueryLoki(t *testing.T) {
	testData := []struct {
		name         string
		sleepAtPage  int
		buildEntries func() []loki.Entry
		buildWant    func() []int
	}{
		{
			name:         "no logs to return",
			buildEntries: func() []loki.Entry { return nil },
			buildWant:    func() []int { return nil },
		},
		{
			name: "all logs",
			buildEntries: func() []loki.Entry {
				var entries []loki.Entry
				for i := -99; i <= 0; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: time.Now().Add(time.Duration(-1) * time.Second),
						Line:      fmt.Sprintf("%v", i),
					})
				}
				return entries
			},
			buildWant: func() []int {
				var want []int
				for i := -99; i <= 0; i++ {
					want = append(want, i)
				}
				return want
			},
		},
		{
			name: "too many logs",
			buildEntries: func() []loki.Entry {
				start := time.Now()
				var entries []loki.Entry
				for i := -49999; i <= 0; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: start.Add(time.Duration(i) * time.Second),
						Line:      fmt.Sprintf("%v", i),
					})
				}
				return entries
			},
			buildWant: func() []int {
				var want []int
				for i := -29999; i <= 0; i++ {
					want = append(want, i)
				}
				return want
			},
		},
		{
			name: "big chunk of unchanging timestamps",
			buildEntries: func() []loki.Entry {
				start := time.Now()
				var entries []loki.Entry
				entries = append(entries, loki.Entry{
					Timestamp: start.Add(-2 * time.Second),
					Line:      "-2",
				})
				for i := 0; i < 40000; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: start.Add(-1 * time.Second),
						Line:      "-1",
					})
				}
				entries = append(entries, loki.Entry{
					Timestamp: start,
					Line:      "0",
				})
				return entries
			},
			buildWant: func() []int {
				var want []int
				want = append(want, -2)
				for i := 0; i < serverMaxLogs-1; i++ {
					want = append(want, -1)
				}
				want = append(want, 0)
				return want
			},
		},
		{
			name: "big chunk of unchanging timestamps at chunk boundary",
			buildEntries: func() []loki.Entry {
				start := time.Now()
				var entries []loki.Entry
				entries = append(entries, loki.Entry{
					Timestamp: start.Add(-2 * time.Second),
					Line:      "-2",
				})
				for i := 0; i < 40000; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: start.Add(-1 * time.Second),
						Line:      "-1",
					})
				}
				for i := 0; i < serverMaxLogs; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: start,
						Line:      "0",
					})
				}
				return entries
			},
			buildWant: func() []int {
				var want []int
				want = append(want, -2)
				for i := 0; i < serverMaxLogs; i++ {
					want = append(want, -1)
				}
				for i := 0; i < serverMaxLogs; i++ {
					want = append(want, 0)
				}
				return want
			},
		},
		{
			name:        "timeout on 1st page",
			sleepAtPage: 1,
			buildEntries: func() []loki.Entry {
				var entries []loki.Entry
				for i := -10000; i <= 0; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: time.Now().Add(time.Duration(-1) * time.Second),
						Line:      fmt.Sprintf("%v", i),
					})
				}
				return entries
			},
			buildWant: func() []int {
				// server should've timed out right away, so no results
				return nil
			},
		},
		{
			name:        "timeout on 2nd page",
			sleepAtPage: 2,
			buildEntries: func() []loki.Entry {
				var entries []loki.Entry
				for i := -10000; i <= 0; i++ {
					entries = append(entries, loki.Entry{
						Timestamp: time.Now().Add(time.Duration(-1) * time.Second),
						Line:      fmt.Sprintf("%v", i),
					})
				}
				return entries
			},
			buildWant: func() []int {
				// expect only the first page due to server timing out
				var want []int
				for i := -999; i <= 0; i++ {
					want = append(want, i)
				}
				return want
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			entries := test.buildEntries()
			want := test.buildWant()

			s := httptest.NewServer(&lokiutil.FakeServer{
				Entries:     entries,
				SleepAtPage: test.sleepAtPage,
			})
			defer s.Close()
			d := &debugServer{
				env: Env{
					GetLokiClient: func() (*loki.Client, error) {
						return &loki.Client{Address: s.URL}, nil
					},
				},
			}

			var got []int
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			out, err := d.queryLoki(ctx, `{foo="bar"}`, 30000)
			if err != nil {
				t.Fatalf("query loki: %v", err)
			}
			for i, l := range out {
				x, err := strconv.ParseInt(l.Entry.Line, 10, 64)
				if err != nil {
					t.Errorf("parse log line %d: %v", i, err)
				}
				got = append(got, int(x))
			}

			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf(`result differs:
          first |   last |    len
       +--------+--------+-------
   got | %6d | %6d | %6d
  want | %6d | %6d | %6d

slice samples:
   got: %v ... %v
  want: %v ... %v`,
					got[0], got[len(got)-1], len(got),
					want[0], want[len(want)-1], len(want),
					got[:4], got[len(got)-4:],
					want[:4], want[len(want)-4:])

				if testing.Verbose() {
					t.Logf("returned lines:\n%v", diff)
				}
			}
		})
	}
}

func TestQuoteLogQL(t *testing.T) {
	for s, q := range map[string]string{
		"abc":  `"abc"`,
		"a'bc": `"a'bc"`,
		`a"bc`: `"a\"bc"`,
		`a
bc`: `"a\nbc"`,
		`ßþ…`: `"ßþ…"`,
	} {
		if quoteLogQLStreamSelector(s) != q {
			t.Errorf("expected quoteLogQL(%q) = %q; got %q", s, q, quoteLogQLStreamSelector(s))
		}
	}
}

func TestHelmReleases(t *testing.T) {
	ctx := pctx.TestContext(t)

	// Build values of secrets, compressed and uncompressed.
	releaseData := map[string]any{
		"config": map[string]any{
			"string": "hello",
			"bool":   false,
		},
		"manifest": `---
apiVersion: v1
kind: Thing
metadata:
    name: thing
---
apiVersion: v1
kind: Thingie
metadata:
    name: thingie
`,
	}
	release, err := json.Marshal(releaseData)
	if err != nil {
		t.Fatalf("marshal reference release: %v", err)
	}
	enc := base64.StdEncoding
	releaseBytes := make([]byte, enc.EncodedLen(len(release)))
	enc.Encode(releaseBytes, release)

	compressed := new(bytes.Buffer)
	gz := gzip.NewWriter(compressed)
	if _, err := gz.Write(release); err != nil {
		t.Fatalf("gzip: %v", err)
	}
	if err := gz.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	releaseBytesCompressed := make([]byte, enc.EncodedLen(compressed.Len()))
	enc.Encode(releaseBytesCompressed, compressed.Bytes())

	// Build a debug server connected to fake k8s that contains a few valid and invalid sample
	// secrets.
	s := &debugServer{
		env: Env{
			GetKubeClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(
					&v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "secret",
							Namespace: "default",
						},
						Data: map[string][]byte{
							"super-secret": []byte("not helm"),
						},
					},
					&v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "some-mistagged-secret",
							Namespace: "default",
							Labels: map[string]string{
								"owner": "helm",
							},
						},
						Data: map[string][]byte{
							"release": []byte("pure junk"),
						},
						Type: "helm.sh/release.v1",
					},
					&v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "sh.helm.release.v1.pachyderm.v9",
							Namespace: "default",
							Labels: map[string]string{
								"owner": "helm",
							},
						},
						Data: map[string][]byte{
							"release": releaseBytes,
						},
						Type: "helm.sh/release.v1",
					},
					&v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "sh.helm.release.v1.pachyderm.v10",
							Namespace: "default",
							Labels: map[string]string{
								"owner": "helm",
							},
						},
						Data: map[string][]byte{
							"release": releaseBytesCompressed,
						},
						Type: "helm.sh/release.v1",
					},
					&v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "sh.helm.release.v2.pachyderm.v11",
							Namespace: "default",
							Labels: map[string]string{
								"owner": "helm",
							},
						},
						Data: map[string][]byte{
							"release": []byte("some junk we don't understand"),
						},
						Type: "helm.sh/release.v2",
					},
				)
			},
			Config: pachconfig.Configuration{
				GlobalConfiguration: &pachconfig.GlobalConfiguration{
					Namespace: "default",
				},
			},
		},
	}

	// Build the debug dump .tar file with just helm release data.
	got := new(bytes.Buffer)
	if err := writeTar(ctx, got, func(ctx context.Context, dfs DumpFS) error {
		return s.collectHelm(ctx, dfs, nil)
	}); err != nil {
		t.Fatalf("helmReleases: %v", err)
	}
	// Iterate over the debug dump, comparing the content of generated files with the reference.
	wantFiles := map[string]any{
		"helm/some-mistagged-secret/metadata.json": map[string]any{
			"name":              "some-mistagged-secret",
			"namespace":         "default",
			"creationTimestamp": nil,
			"labels": map[string]any{
				"owner": "helm",
			},
		},
		"helm/some-mistagged-secret/release/error.txt": "decode base64: illegal base64 data at input byte 4\n",
		"helm/sh.helm.release.v1.pachyderm.v9/metadata.json": map[string]any{
			"name":              "sh.helm.release.v1.pachyderm.v9",
			"namespace":         "default",
			"creationTimestamp": nil,
			"labels": map[string]any{
				"owner": "helm",
			},
		},
		"helm/sh.helm.release.v1.pachyderm.v9/release.json":  releaseData,
		"helm/sh.helm.release.v1.pachyderm.v9/manifest.yaml": releaseData["manifest"].(string),
		"helm/sh.helm.release.v1.pachyderm.v9/values.yaml":   releaseData["config"].(map[string]any),
		"helm/sh.helm.release.v1.pachyderm.v10/metadata.json": map[string]any{
			"name":              "sh.helm.release.v1.pachyderm.v10",
			"namespace":         "default",
			"creationTimestamp": nil,
			"labels": map[string]any{
				"owner": "helm",
			},
		},
		"helm/sh.helm.release.v1.pachyderm.v10/release.json":  releaseData,
		"helm/sh.helm.release.v1.pachyderm.v10/manifest.yaml": releaseData["manifest"].(string),
		"helm/sh.helm.release.v1.pachyderm.v10/values.yaml":   releaseData["config"].(map[string]any),
		"helm/sh.helm.release.v2.pachyderm.v11/error.txt":     "helm-owned secret of unknown version; got helm.sh/release.v2 want helm.sh/release.v1\n",
	}
	if err := tarutil.Iterate(got, func(f tarutil.File) error {
		// Extract the content of the file.
		buf := new(bytes.Buffer)
		if err := f.Content(buf); err != nil {
			return errors.Wrap(err, "get content")
		}
		// Extract the header of the file and make sure it's a file we expect.
		h, err := f.Header()
		if err != nil {
			return errors.Wrap(err, "get header")
		}
		want, ok := wantFiles[h.Name]
		if !ok {
			t.Errorf("unexpected file %v (content: %v)", h.Name, buf.String())
			return nil
		}
		delete(wantFiles, h.Name)

		// Transform the content into an object if appropriate.
		var got any
		if strings.HasSuffix(h.Name, ".json") {
			var x map[string]any
			if err := json.Unmarshal(buf.Bytes(), &x); err != nil {
				return errors.Wrapf(err, "%v: unmarshal json", h.Name)
			}
			got = x
		} else if strings.HasSuffix(h.Name, "/values.yaml") {
			// We only unmarshal values.yaml because manifest.yaml should be preserved
			// verbatim from the value stored in the secret; no sort order to worry
			// about when doing a string comparison.  values.yaml can be sorted
			// arbitrarily, however.
			var x map[string]any
			if err := yaml.Unmarshal(buf.Bytes(), &x); err != nil {
				return errors.Wrapf(err, "%v: unmarshal yaml", h.Name)
			}
			got = x
		} else {
			got = buf.String()
		}

		// Diff the (unmarshaled) content with the reference value.
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("content of %v (+got -want):\n%s", h.Name, diff)
			return nil
		}

		return nil
	}); err != nil {
		t.Fatalf("Iterate: %v", err)
	}

	// Check that we saw all the files we expected.
	for f := range wantFiles {
		t.Errorf("did not see expected file %v", f)
	}
}

func TestLoadTestEmbed(t *testing.T) {
	for _, s := range defaultLoadSpecs {
		require.NotEqual(t, "", s)
	}
}

func TestListApps(t *testing.T) {
	ctx := pctx.TestContext(t)
	s := &debugServer{
		env: Env{
			GetKubeClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "etcd-0",
							Namespace: "default",
							Labels: map[string]string{
								"suite": "pachyderm",
								"app":   "etcd",
							},
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "etcd",
								},
							},
						},
						Status: v1.PodStatus{
							PodIP: "10.0.0.2",
						},
					},
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "etcd-1",
							Namespace: "default",
							Labels: map[string]string{
								"suite": "pachyderm",
								"app":   "etcd",
							},
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "etcd",
								},
							},
						},
						Status: v1.PodStatus{
							PodIP: "10.0.0.3",
						},
					},
					&v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "default-edges-abc123",
							Namespace: "default",
							Labels: map[string]string{
								"suite":           "pachyderm",
								"app":             "pipeline",
								"pipelineProject": "default",
								"pipelineName":    "edges",
							},
						},
						Spec: v1.PodSpec{
							InitContainers: []v1.Container{
								{
									Name: "init",
								},
							},
							Containers: []v1.Container{
								{
									Name: "user",
								},
								{
									Name: "storage",
								},
							},
						},
						Status: v1.PodStatus{
							PodIP: "10.0.0.4",
						},
					},
				)
			},
			Config: pachconfig.Configuration{
				GlobalConfiguration: &pachconfig.GlobalConfiguration{
					Namespace: "default",
				},
			},
		},
	}
	gotRunning, gotPossible, err := s.listApps(ctx, []*pps.Pipeline{
		{Project: &pfs.Project{Name: "default"}, Name: "edges"},
		{Project: &pfs.Project{Name: "default"}, Name: "montage"},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantPossible := []*debug.App{
		{
			Name: "default/edges",
			Pipeline: &debug.Pipeline{
				Project: "default",
				Name:    "edges",
			},
			Pods: []*debug.Pod{
				{
					Name:       "default-edges-abc123",
					Ip:         "10.0.0.4",
					Containers: []string{"user", "storage"},
				},
			},
		},
		{
			Name: "default/montage",
			Pipeline: &debug.Pipeline{
				Project: "default",
				Name:    "montage",
			},
		},
		{
			Name: "etcd",
			Pods: []*debug.Pod{
				{
					Name:       "etcd-0",
					Ip:         "10.0.0.2",
					Containers: []string{"etcd"},
				},
				{
					Name:       "etcd-1",
					Ip:         "10.0.0.3",
					Containers: []string{"etcd"},
				},
			},
		},
	}
	if diff := cmp.Diff(wantPossible, gotPossible, protocmp.Transform()); diff != "" {
		t.Errorf("possible apps (-want +got):\n%s", diff)
	}

	wantRunning := []*debug.App{wantPossible[0], wantPossible[2]}
	if diff := cmp.Diff(wantRunning, gotRunning, protocmp.Transform()); diff != "" {
		t.Errorf("running apps (-want +got):\n%s", diff)
	}
}
