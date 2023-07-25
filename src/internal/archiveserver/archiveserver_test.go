package archiveserver

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type fakePFS struct {
	pfs.APIClient
}

type getFileTARClient struct {
	pfs.API_GetFileTARClient
	recvd bool
	files map[string]string
}

func (c *getFileTARClient) Recv() (*wrapperspb.BytesValue, error) {
	if !c.recvd {
		c.recvd = true
		buf := new(bytes.Buffer)
		w := tar.NewWriter(buf)
		for file, content := range c.files {
			if file == "error" {
				return nil, errors.New(content)
			}
			if err := w.WriteHeader(&tar.Header{
				Name: file,
				Size: int64(len(content)),
			}); err != nil {
				return nil, errors.Wrap(err, "WriteHeader")
			}
			if _, err := fmt.Fprintf(w, "%s", content); err != nil {
				return nil, errors.Wrapf(err, "Write(%s)", file)
			}
		}
		if err := w.Close(); err != nil {
			return nil, errors.Wrap(err, "Close")
		}
		return wrapperspb.Bytes(buf.Bytes()), nil
	}
	return nil, io.EOF
}

const fakeCommit = "44444444444444444444444444444444"

func (fakePFS) GetFileTAR(ctx context.Context, req *pfs.GetFileRequest, opts ...grpc.CallOption) (pfs.API_GetFileTARClient, error) {
	log.Debug(ctx, "GetFileTAR", log.Proto("request", req))
	files := map[string]string{}
	if req.File.Commit.Id == fakeCommit {
		if req.File.Commit.Repo.Name == "images" && req.File.Commit.Repo.Project.Name == "default" && req.File.Path == "/" {
			files["/hello.txt"] = "hello"
			files["/a/"] = ""
			files["/a/nested/"] = ""
			files["/a/nested/file.txt"] = "i'm nested!"
		}
		if req.File.Commit.Repo.Name == "images" && req.File.Commit.Repo.Project.Name == "default" && req.File.Path == "/hello.txt" {
			files["/hello.txt"] = "hello"
		}
		if req.File.Commit.Repo.Name == "montage" && req.File.Commit.Repo.Project.Name == "default" && req.File.Path == "/montage.png" {
			files["/montage.png"] = "beautiful artwork is here"
		}
		if req.File.Path == "/error.txt" {
			files["error"] = "error reading from the server"
		}
	}
	return &getFileTARClient{files: files}, nil
}

func (fakePFS) InspectBranch(ctx context.Context, req *pfs.InspectBranchRequest, opts ...grpc.CallOption) (*pfs.BranchInfo, error) {
	log.Debug(ctx, "InspectBranch", log.Proto("request", req))
	if req.Branch.Name == "master" {
		return &pfs.BranchInfo{
			Head: &pfs.Commit{
				Repo: req.Branch.Repo,
				Id:   fakeCommit,
			},
		}, nil
	}
	return nil, errors.New("not found")
}

// so TestHTTP and FuzzHTTP can share the implementation
func doTest(t *testing.T, method, url string) (int, *bytes.Buffer) {
	fake := &client.APIClient{}
	fake.PfsAPIClient = &fakePFS{}

	ctx := pctx.TestContext(t)
	s := NewHTTP(0, func(ctx context.Context) *client.APIClient { return fake.WithCtx(ctx) })

	req := httptest.NewRequest(method, url, nil)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	return rec.Code, rec.Body
}

// These URLs should be fetched by both TestHTTP (deep validation) and FuzzHTTP (good starting
// points for fuzzing).
var testData = []struct {
	name      string
	method    string
	url       string
	wantCode  int
	wantFiles map[string]string
}{
	{
		name:     "unknown route",
		method:   "GET",
		url:      "http://pachyderm.example.com/what-is-this?",
		wantCode: http.StatusNotFound,
	},
	{
		name:     "health",
		method:   "GET",
		url:      "http://pachyderm.example.com/healthz",
		wantCode: http.StatusOK,
	},
	{
		name:      "empty download",
		method:    "GET",
		url:       "http://pachyderm.example.com/archive/AQ.zip",
		wantCode:  http.StatusOK,
		wantFiles: map[string]string{},
	},
	{
		name:      "empty download with auth token",
		method:    "GET",
		url:       "http://pachyderm.example.com/archive/AQ.zip?authn-token=foobar",
		wantCode:  http.StatusOK,
		wantFiles: map[string]string{},
	},
	{
		name:     "invalid output format",
		method:   "GET",
		url:      "http://pachyderm.example.com/archive/AQ.tar.bz2",
		wantCode: http.StatusBadRequest,
	},
	{
		name:     "unknown method",
		method:   "HEAD",
		url:      "http://pachyderm.example.com/archive/AQ.zip",
		wantCode: http.StatusMethodNotAllowed,
	},
	{
		name:     "download with some content",
		method:   "GET",
		url:      "https://pachyderm.example.com/archive/ASi1L_0EaHUBAEQCZGVmYXVsdC9pbWFnZXNAbWFzdGVyOi8AbW9udGFnZS5wbmcAAxQEBQPYsGPLbFDb.zip",
		wantCode: http.StatusOK,
		wantFiles: map[string]string{
			"default/images/44444444444444444444444444444444/hello.txt":         "hello",
			"default/images/44444444444444444444444444444444/a/nested/file.txt": "i'm nested!",
			"default/montage/44444444444444444444444444444444/montage.png":      "beautiful artwork is here",
		},
	},
	{
		name:     "download with some content, commit references in URL",
		method:   "GET",
		url:      "https://pachyderm.example.com/archive/ASi1L_0EaL0BAIQCZGVmYXVsdC9pbWFnZXNANDovaGVsbG8udHh0AG1vbnRhZ2UucG5nAAQATRHgK2e8IpIGLAGgJI8S.zip",
		wantCode: http.StatusOK,
		wantFiles: map[string]string{
			"default/images/44444444444444444444444444444444/hello.txt":    "hello",
			"default/montage/44444444444444444444444444444444/montage.png": "beautiful artwork is here",
		},
	},
	{
		name:     "download with an error reading files",
		method:   "GET",
		url:      "https://pachyderm.example.com/archive/ASi1L_0EaPkAAGRlZmF1bHQvdGVzdEBtYXN0ZXI6L2Vycm9yLnR4dABwDhIY.zip",
		wantCode: http.StatusOK,
		wantFiles: map[string]string{
			"@error.txt": "path default/test@=44444444444444444444444444444444:/error.txt: read TAR header: error reading from the server\n",
		},
	},
}

func TestHTTP(t *testing.T) {
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			code, body := doTest(t, test.method, test.url)
			if got, want := code, test.wantCode; got != want {
				t.Errorf("response code:\n  got: %v\n want: %v", got, want)
			}

			if test.wantFiles != nil {
				bs := body.Bytes() // zip needs a ReaderAt, which Body isn't.
				r, err := zip.NewReader(bytes.NewReader(bs), int64(len(bs)))
				if err != nil {
					t.Fatalf("create zip reader: %v", err)
				}
				got := map[string]string{}
				for _, fileinfo := range r.File {
					file, err := r.Open(fileinfo.Name)
					if err != nil {
						t.Fatalf("open %v: %v", file, err)
					}
					buf := new(bytes.Buffer)
					if _, err := io.Copy(buf, file); err != nil {
						t.Fatalf("read %v: %v", file, err)
					}
					got[fileinfo.Name] = buf.String()
				}
				if diff := cmp.Diff(got, test.wantFiles); diff != "" {
					t.Errorf("downloaded files (-got +want):\n%s", diff)
				}
			}
		})
	}
}

func FuzzHTTP(f *testing.F) {
	for _, test := range testData {
		f.Add(test.url)
		f.Log(test.url)
	}
	f.Fuzz(func(t *testing.T, u string) {
		// Skip the invalid URLs it generates.
		if _, err := http.ReadRequest(bufio.NewReader(strings.NewReader("GET " + u + " HTTP/1.0\r\n\r\n"))); err != nil {
			return
		}

		up, err := url.Parse(u)
		isArchive := err == nil && strings.HasPrefix(up.Path, "/archive/")

		// Then use the normal testing machinery.
		code, body := doTest(t, "GET", u)
		if code == http.StatusOK && isArchive && body.Len() > 0 {
			// If the code is OK and the URL starts with /archive/, then there should
			// be either nothing, or a zip.  Assert that the ZIP is readable.
			bs := body.Bytes()
			t.Logf("potential zip bytes: %x %s", bs, bs)

			r, err := zip.NewReader(bytes.NewReader(bs), int64(len(bs)))
			if err != nil {
				t.Fatalf("zip.NewReader: %v", err)
			}
			for _, fileinfo := range r.File {
				file, err := r.Open(fileinfo.Name)
				if err != nil {
					t.Fatalf("open(%v): %v", fileinfo.Name, err)
				}
				if _, err := io.Copy(io.Discard, file); err != nil {
					t.Fatalf("read error(%v): %v", fileinfo.Name, err)
				}
			}
		}
	})
}
