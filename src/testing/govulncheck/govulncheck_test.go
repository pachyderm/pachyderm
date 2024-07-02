package govulncheck

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"golang.org/x/exp/maps"
	"golang.org/x/vuln/scan"
)

var (
	// Any packages that we want to scan during tests; these paths should contain a go.mod.
	packages = []string{
		"",
		"examples/spouts/go-rabbitmq-spout/source",
	}
	// Any vulnerabilities to ignore.
	ignore = map[string]struct{}{
		// Using the etcd client at all is a security problem because it might connect to an
		// etcd server with an insecure version of TLS; no fixed available.
		"GO-2024-2527": {},
	}
)

func TestGovulncheck(t *testing.T) {
	// Find the source code.
	mod, err := runfiles.Rlocation("_main/MODULE.bazel")
	if err != nil {
		t.Fatalf("find MODULE.bazel: %v", err)
	}
	real, err := os.Readlink(mod)
	if err != nil {
		t.Fatalf("readlink(MODULE.bazel): %v", err)
	}
	root := filepath.Dir(real)

	// Find go.
	if goBin, err := runfiles.Rlocation("rules_go~~go_sdk~go_sdk/bin/go"); err != nil {
		t.Logf("can't find go binary; using installed go instead: %v", err)
	} else {
		path := os.Getenv("PATH")
		path = filepath.Dir(goBin) + ":" + path
		os.Setenv("PATH", path)
	}

	// govulncheck ends up shelling out to Go and downloading modules, etc.  It is slow to
	// delete them all when the job is done (as would happen with t.TempDir), so we don't
	// bother.
	os.Setenv("GOPATH", "/tmp/govulncheck/gopath")
	os.Setenv("GOCACHE", "/tmp/govulncheck/gocache")

	// Run govulncheck for each workspace-relative package listed above.
	for _, pkg := range packages {
		name := pkg
		if name == "" {
			name = "main module"
		}
		t.Run("scan "+name, func(t *testing.T) {
			govulncheck(t, filepath.Join(root, pkg))
		})
	}
}

type message struct {
	Progress *struct {
		Message string `json:"message"`
	} `json:"progress"`
	Finding *struct {
		OSV   string `json:"osv"`
		Trace []struct {
			Function string `json:"function"` // If this is set, we call the vulnerable code at some point.
		} `json:"trace"`
	}
}

func decode(t *testing.T, r io.Reader) ([]string, *bytes.Buffer, error) {
	t.Helper()
	processed := new(bytes.Buffer) // processed contains the JSON to reformat with govulncheck -mode=convert
	vulnerable := make(map[string]struct{})
	ignored := make(map[string]struct{})

	d := json.NewDecoder(r)
	for {
		var x json.RawMessage
		if err := d.Decode(&x); err != nil {
			if errors.Is(err, io.EOF) {
				result := maps.Keys(vulnerable)
				sort.Strings(result)
				return result, processed, nil
			}
			return nil, nil, errors.Wrap(err, "read json")
		}
		var m message
		if err := json.Unmarshal([]byte(x), &m); err != nil {
			return nil, nil, errors.Wrap(err, "interpret message")
		}
		switch {
		case m.Progress != nil:
			t.Log(m.Progress.Message)
		case m.Finding != nil:
			f := m.Finding
			if len(f.Trace) > 0 && f.Trace[0].Function != "" {
				if _, ok := ignore[f.OSV]; ok {
					if _, ok := ignored[f.OSV]; !ok {
						ignored[f.OSV] = struct{}{}
						t.Logf("ignoring vulnerability %v", f.OSV)
					}
					break
				}
				if _, ok := vulnerable[f.OSV]; !ok {
					t.Logf("%v identified as a vulnerability; further output follows", f.OSV)
					vulnerable[f.OSV] = struct{}{}
				}
			}
			fallthrough
		default: // Config and OSV message types.
			processed.Write([]byte(x))
		}
	}
}

func govulncheck(t *testing.T, pkg string) {
	t.Helper()
	ctx, cancel := context.WithCancelCause(pctx.TestContext(t))
	doneCh := make(chan struct{})
	cmd := scan.Command(ctx, "-C", pkg, "-json", "./...")
	r, w := io.Pipe()
	cmd.Stdout = w
	go func() {
		defer w.Close()
		if err := cmd.Start(); err != nil {
			cancel(errors.Wrap(err, "start govulncheck scan"))
		}
		if err := cmd.Wait(); err != nil {
			cancel(errors.Wrap(err, "await govulncheck scan result"))
		}
		close(doneCh)
	}()
	vulnerable, processed, err := decode(t, r)
	if err != nil {
		t.Fatalf("decode vulnerability report: %v", err)
	}
	select {
	case <-ctx.Done():
		if err := context.Cause(ctx); err != nil {
			t.Fatalf("scan for vulnerabilities: %v", err)
		}
	case <-doneCh:
		t.Log("Scan done; formatting details...")
	}

	ctx = pctx.TestContext(t)
	cmd = scan.Command(ctx, "-mode=convert")
	cmd.Stdin = processed
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start conversion command: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait for conversion command: %v", err)
	}

	if len(vulnerable) > 0 {
		t.Fatalf("code is vulnerable to: %v; see above output for details", strings.Join(vulnerable, ", "))
	}
}

func TestDecode(t *testing.T) {
	fh, err := os.Open("testdata/ignored.json")
	if err != nil {
		t.Fatalf("open testdata: %v", err)
	}
	t.Cleanup(func() { fh.Close() })
	got, _, err := decode(t, fh)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := []string{"GO-2024-2512"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("vulnerabilities in testdata report (-want +got):\n%s", diff)
	}
}
