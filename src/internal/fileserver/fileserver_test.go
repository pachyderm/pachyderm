package fileserver_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/constants"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmputil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/fileserver"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestDownload(t *testing.T) {
	// Setup a PFS server.
	rctx := pctx.TestContext(t)
	e := realenv.NewRealEnvWithIdentity(rctx, t, dockertestenv.NewTestDBConfig(t))
	// Fileserver for testing.
	s := &fileserver.Server{
		ClientFactory: func(ctx context.Context) *client.APIClient {
			c := e.PachClient.WithCtx(ctx)
			c.SetAuthToken("")
			return c
		},
	}

	// Add test data.  There will be one repo, "default/test", with a branch "master".  "master"
	// will have two commits, one finished and one open.
	r := &pfs.Repo{
		Name: "test",
		Type: pfs.UserRepoType,
		Project: &pfs.Project{
			Name: pfs.DefaultProjectName,
		},
	}
	master := &pfs.Branch{
		Name: "master",
		Repo: r,
	}
	if _, err := e.PachClient.PfsAPIClient.CreateRepo(rctx, &pfs.CreateRepoRequest{
		Repo: r,
	}); err != nil {
		t.Fatalf("create test repo: %v", err)
	}
	// Create a commit that will be finished by the time the tests run.
	finishedCommit, err := e.PachClient.PfsAPIClient.StartCommit(rctx, &pfs.StartCommitRequest{
		Branch: master,
	})
	if err != nil {
		t.Fatalf("start commit ('finished commit'): %v", err)
	}
	mfc, err := e.PachClient.PfsAPIClient.ModifyFile(rctx)
	if err != nil {
		t.Fatalf("start modifyfile ('finished commit'): %v", err)
	}
	if err := mfc.Send(&pfs.ModifyFileRequest{
		Body: &pfs.ModifyFileRequest_SetCommit{
			SetCommit: finishedCommit,
		},
	}); err != nil {
		t.Fatalf("modify file ('finished commit'): set commit (finished commit): %v", err)
	}
	if err := mfc.Send(&pfs.ModifyFileRequest{
		Body: &pfs.ModifyFileRequest_AddFile{
			AddFile: &pfs.AddFile{
				Path: "/sub/directory/test.txt",
				Source: &pfs.AddFile_Raw{
					Raw: &wrapperspb.BytesValue{
						Value: []byte("hello, world\n"),
					},
				},
			},
		},
	}); err != nil {
		t.Fatalf("modify file ('finished commit'): add /sub/directory/test.txt: %v", err)
	}
	// Create a "big" (26KB) file with predictable data; will be used to test HTTP range
	// requests.
	bigData := new(bytes.Buffer)
	for i := 'A'; i <= 'Z'; i++ {
		for j := 0; j < 1000; j++ {
			bigData.WriteRune(i)
		}
	}
	if err := mfc.Send(&pfs.ModifyFileRequest{
		Body: &pfs.ModifyFileRequest_AddFile{
			AddFile: &pfs.AddFile{
				Path: "big.txt",
				Source: &pfs.AddFile_Raw{
					Raw: &wrapperspb.BytesValue{
						Value: bigData.Bytes(),
					},
				},
			},
		},
	}); err != nil {
		t.Fatalf("modify file ('finished commit'): add big.txt: %v", err)
	}
	if _, err := mfc.CloseAndRecv(); err != nil {
		t.Fatalf("modify file ('finished commit'): CloseAndRecv: %v", err)
	}
	if _, err := e.PachClient.PfsAPIClient.FinishCommit(rctx, &pfs.FinishCommitRequest{
		Commit: finishedCommit,
	}); err != nil {
		t.Fatalf("finish commit ('finished commit'): %v", err)
	}
	finishedInfo, err := e.PachClient.PfsAPIClient.InspectCommit(rctx, &pfs.InspectCommitRequest{
		Commit: finishedCommit,
		Wait:   pfs.CommitState_FINISHED, // This is not true immediately after FinishCommit returns.
	})
	if err != nil {
		t.Fatalf("inspect commit ('finished commit'): %v", err)
	}

	// File times ("Committed") are based on the finishing time, not the finished time.
	finishedAt := finishedInfo.GetFinishing().AsTime().In(time.UTC)

	// Create a branch reference to the finished commit; so that the tests can distinguish
	// between branch-head-open vs. branch-head-closed reads.  These have the same caching
	// semantics but are different cases.  A closed branch head is un-cacheable because the
	// branch head can move.  An open branch head is un-cacheable because the commit content can
	// change.
	if _, err := e.PachClient.PfsAPIClient.CreateBranch(rctx, &pfs.CreateBranchRequest{
		Head: finishedCommit,
		Branch: &pfs.Branch{
			Repo: r,
			Name: "done",
		},
	}); err != nil {
		t.Fatalf("create branch done: %v", err)
	}

	openCommit, err := e.PachClient.PfsAPIClient.StartCommit(rctx, &pfs.StartCommitRequest{
		Branch: master,
	})
	if err != nil {
		t.Fatalf("start commit ('open commit'): %v", err)
	}
	mfc, err = e.PachClient.PfsAPIClient.ModifyFile(rctx)
	if err != nil {
		t.Fatalf("modify file ('open commit'): %v", err)
	}
	if err := mfc.Send(&pfs.ModifyFileRequest{
		Body: &pfs.ModifyFileRequest_SetCommit{
			SetCommit: openCommit,
		},
	}); err != nil {
		t.Fatalf("modify file ('open commit'): set commit: %v", err)
	}
	if err := mfc.Send(&pfs.ModifyFileRequest{
		Body: &pfs.ModifyFileRequest_DeleteFile{
			DeleteFile: &pfs.DeleteFile{
				Path: "/sub/directory/test.txt",
			},
		},
	}); err != nil {
		t.Fatalf("modify file ('open commit'): delete /sub/directory/test.txt: %v", err)
	}
	if err := mfc.Send(&pfs.ModifyFileRequest{
		Body: &pfs.ModifyFileRequest_AddFile{
			AddFile: &pfs.AddFile{
				Path: "/sub/directory/test.txt",
				Source: &pfs.AddFile_Raw{
					Raw: &wrapperspb.BytesValue{
						Value: []byte("goodbye, world\n"),
					},
				},
			},
		},
	}); err != nil {
		t.Fatalf("modify file ('open commit'): add new /sub/directory/test.txt: %v", err)
	}
	if _, err := mfc.CloseAndRecv(); err != nil {
		t.Fatalf("modify file ('open commit'): CloseAndRecv: %v", err)
	}
	testutil.ActivateAuthClient(t, e.PachClient, strconv.Itoa(int(e.ServiceEnv.Config().PeerPort)))

	c := e.PachClient.WithCtx(rctx)
	c.SetAuthToken(testutil.RootToken)
	token, err := c.GetRobotToken(c.Ctx(), &auth.GetRobotTokenRequest{
		Robot: "alice",
	})
	if err != nil {
		t.Fatalf("create auth token: %v", err)
	}
	if _, err := c.ModifyRoleBinding(c.Ctx(), &auth.ModifyRoleBindingRequest{
		Principal: "robot:alice",
		Roles:     []string{auth.RepoReaderRole},
		Resource: &auth.Resource{
			Type: auth.ResourceType_REPO,
			Name: "default/test",
		},
	}); err != nil {
		t.Fatalf("give alice reader on default/test: %v", err)
	}

	// The actual test starts here.
	testData := []struct {
		name          string
		method        string
		url           string
		requestHeader http.Header
		wantCode      int
		wantContent   string
		wantHeader    http.Header
	}{
		{
			name:        "non-pfs URL",
			method:      http.MethodGet,
			url:         "https://example.com/notpfs/default/test/master/test.txt",
			wantCode:    http.StatusInternalServerError,
			wantContent: "unexpectedly handling request not for /pfs; got notpfs want pfs",
		},
		{
			name:        "directory redirect",
			method:      http.MethodGet,
			url:         "https://example.com/pfs/default/test/master/sub/directory",
			wantCode:    http.StatusMovedPermanently,
			wantContent: "<a href=\"/pfs/default/test/master/sub/directory/\">Moved Permanently</a>.\n\n",
			wantHeader: http.Header{
				"Cache-Control": {"private, no-cache"},
				"Content-Type":  {"text/html; charset=utf-8"},
				"Etag":          {`"f103067f40b04b3e6e5ad1dbf053d0c68454e623c200a63e400a60de84ad632c"`},
				"Location":      {"/pfs/default/test/master/sub/directory/"},
				"Vary":          {"authn-token"},
			},
		},
		{
			name:        "get nonexistent file",
			method:      http.MethodGet,
			url:         fmt.Sprintf("https://example.com/pfs/default/test/%v/nonexist.txt", finishedCommit.Id),
			wantCode:    http.StatusNotFound,
			wantContent: "problem inspecting file: rpc error: code = NotFound desc = file nonexist.txt not found in repo default/test at commit " + finishedCommit.Id,
		},
		{
			name:        "get master test.txt",
			method:      http.MethodGet,
			url:         "https://example.com/pfs/default/test/master/sub/directory/test.txt",
			wantCode:    http.StatusOK,
			wantContent: "goodbye, world\n",
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private, no-cache"}, // no-cache because "master"
				"Content-Length": {"15"},
				"Etag":           {`"7f25604c8f64d4e40377c006dcaa47626e4b1d93b09f1f8252e14e643c8e8f02"`},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:     "head master test.txt",
			method:   http.MethodHead,
			url:      "https://example.com/pfs/default/test/master/sub/directory/test.txt",
			wantCode: http.StatusOK,
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private, no-cache"},
				"Content-Length": {"15"},
				"Etag":           {`"7f25604c8f64d4e40377c006dcaa47626e4b1d93b09f1f8252e14e643c8e8f02"`},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:        "get open test.txt",
			method:      http.MethodGet,
			url:         fmt.Sprintf("https://example.com/pfs/default/test/%v/sub/directory/test.txt", openCommit.Id),
			wantCode:    http.StatusOK,
			wantContent: "goodbye, world\n",
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private, no-cache"}, // no-cache because open commit
				"Content-Length": {"15"},
				"Etag":           {`"7f25604c8f64d4e40377c006dcaa47626e4b1d93b09f1f8252e14e643c8e8f02"`},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:     "head open test.txt",
			method:   http.MethodHead,
			url:      fmt.Sprintf("https://example.com/pfs/default/test/%v/sub/directory/test.txt", openCommit.Id),
			wantCode: http.StatusOK,
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private, no-cache"},
				"Content-Length": {"15"},
				"Etag":           {`"7f25604c8f64d4e40377c006dcaa47626e4b1d93b09f1f8252e14e643c8e8f02"`},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:        "get finished test.txt",
			method:      http.MethodGet,
			url:         fmt.Sprintf("https://example.com/pfs/default/test/%v/sub/directory/test.txt", finishedCommit.Id),
			wantCode:    http.StatusOK,
			wantContent: "hello, world\n",
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private"}, // cacheable because finished commit
				"Content-Length": {"13"},
				"Etag":           {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
				"Last-Modified":  {finishedAt.Format(http.TimeFormat)},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:     "head finished test.txt",
			method:   http.MethodHead,
			url:      fmt.Sprintf("https://example.com/pfs/default/test/%v/sub/directory/test.txt", finishedCommit.Id),
			wantCode: http.StatusOK,
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private"}, // cacheable because finished commit
				"Content-Length": {"13"},
				"Etag":           {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
				"Last-Modified":  {finishedAt.Format(http.TimeFormat)},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:        "get done test.txt",
			method:      http.MethodGet,
			url:         "https://example.com/pfs/default/test/done/sub/directory/test.txt",
			wantCode:    http.StatusOK,
			wantContent: "hello, world\n",
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private, no-cache"}, // not cacheable because branch
				"Content-Length": {"13"},
				"Etag":           {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
				"Last-Modified":  {finishedAt.Format(http.TimeFormat)},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:     "head done test.txt",
			method:   http.MethodHead,
			url:      "https://example.com/pfs/default/test/done/sub/directory/test.txt",
			wantCode: http.StatusOK,
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private, no-cache"}, // not cacheable because branch
				"Content-Length": {"13"},
				"Etag":           {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
				"Last-Modified":  {finishedAt.Format(http.TimeFormat)},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:   "if-none-match head test.txt@done, match",
			method: http.MethodHead,
			url:    "https://example.com/pfs/default/test/done/sub/directory/test.txt",
			requestHeader: http.Header{
				"If-None-Match": {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
			},
			wantCode:    http.StatusNotModified,
			wantContent: "",
			wantHeader: http.Header{
				"Accept-Ranges": {"bytes"},
				"Cache-Control": {"private, no-cache"},
				"Etag":          {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
				"Last-Modified": {finishedAt.Format(http.TimeFormat)},
				"Vary":          {"authn-token"},
			},
		},
		{
			name:   "if-none-match get test.txt@done, match",
			method: http.MethodGet,
			url:    "https://example.com/pfs/default/test/done/sub/directory/test.txt",
			requestHeader: http.Header{
				"If-None-Match": {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
			},
			wantCode:    http.StatusNotModified,
			wantContent: "",
			wantHeader: http.Header{
				"Accept-Ranges": {"bytes"},
				"Cache-Control": {"private, no-cache"},
				"Etag":          {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
				"Last-Modified": {finishedAt.Format(http.TimeFormat)},
				"Vary":          {"authn-token"},
			},
		}, {
			name:   "if-none-match head test.txt@done, no match",
			method: http.MethodHead,
			url:    "https://example.com/pfs/default/test/done/sub/directory/test.txt",
			requestHeader: http.Header{
				"If-None-Match": {`"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`},
			},
			wantCode:    http.StatusOK,
			wantContent: "",
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private, no-cache"},
				"Content-Length": {"13"},
				"Etag":           {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
				"Last-Modified":  {finishedAt.Format(http.TimeFormat)},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:   "if-none-match get test.txt@done, no match",
			method: http.MethodGet,
			url:    "https://example.com/pfs/default/test/done/sub/directory/test.txt",
			requestHeader: http.Header{
				"If-None-Match": {`"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`},
			},
			wantCode:    http.StatusOK,
			wantContent: "hello, world\n",
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private, no-cache"},
				"Content-Length": {"13"},
				"Etag":           {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
				"Last-Modified":  {finishedAt.Format(http.TimeFormat)},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:   "if-modified-since get test.txt@done, unmodified",
			method: http.MethodGet,
			url:    "https://example.com/pfs/default/test/done/sub/directory/test.txt",
			requestHeader: http.Header{
				"If-Modified-Since": {finishedAt.Add(time.Second).Format(http.TimeFormat)},
			},
			wantCode:    http.StatusNotModified,
			wantContent: "",
			wantHeader: http.Header{
				"Accept-Ranges": {"bytes"},
				"Cache-Control": {"private, no-cache"},
				"Etag":          {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
				"Last-Modified": {finishedAt.Format(http.TimeFormat)},
				"Vary":          {"authn-token"},
			},
		},
		{
			name:   "if-modified-since get test.txt@done, modified",
			method: http.MethodGet,
			url:    "https://example.com/pfs/default/test/done/sub/directory/test.txt",
			requestHeader: http.Header{
				"If-Modified-Since": {time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Format(http.TimeFormat)},
			},
			wantCode:    http.StatusOK,
			wantContent: "hello, world\n",
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Cache-Control":  {"private, no-cache"},
				"Content-Length": {"13"},
				"Etag":           {`"918cd0e91afb64becb2d77e7cba9d1e8ea15ad5a26c16bbaf629ef916eaeb414"`},
				"Last-Modified":  {finishedAt.Format(http.TimeFormat)},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:        "big.txt, finished",
			method:      http.MethodGet,
			url:         fmt.Sprintf("https://example.com/pfs/default/test/%v/big.txt", finishedCommit.Id),
			wantContent: bigData.String(),
			wantCode:    http.StatusOK,
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Content-Length": {strconv.Itoa(26 * 1000)},
				"Cache-Control":  {"private"},
				"Etag":           {`"ee45bb2661662b371f3100176c518243dbbe282116a05bb8f025a9c23b97ea02"`},
				"Last-Modified":  {finishedAt.Format(http.TimeFormat)},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:   "range download, finished",
			method: http.MethodGet,
			url:    fmt.Sprintf("https://example.com/pfs/default/test/%v/big.txt", finishedCommit.Id),
			requestHeader: http.Header{
				"Range": {"bytes=0-9"},
			},
			wantContent: "AAAAAAAAAA",
			wantCode:    http.StatusPartialContent,
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Content-Length": {"10"},
				"Cache-Control":  {"private"},
				"Etag":           {`"ee45bb2661662b371f3100176c518243dbbe282116a05bb8f025a9c23b97ea02"`},
				"Last-Modified":  {finishedAt.Format(http.TimeFormat)},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:   "range download, finished, if match, matches",
			method: http.MethodGet,
			url:    fmt.Sprintf("https://example.com/pfs/default/test/%v/big.txt", finishedCommit.Id),
			requestHeader: http.Header{
				"Range":    {"bytes=25000-25010"},
				"If-Match": {`"ee45bb2661662b371f3100176c518243dbbe282116a05bb8f025a9c23b97ea02"`},
			},
			wantContent: "ZZZZZZZZZZZ",
			wantCode:    http.StatusPartialContent,
			wantHeader: http.Header{
				"Accept-Ranges":  {"bytes"},
				"Content-Length": {"11"},
				"Cache-Control":  {"private"},
				"Etag":           {`"ee45bb2661662b371f3100176c518243dbbe282116a05bb8f025a9c23b97ea02"`},
				"Last-Modified":  {finishedAt.Format(http.TimeFormat)},
				"Vary":           {"authn-token"},
			},
		},
		{
			name:   "range download, finished, if match, doesn't match",
			method: http.MethodGet,
			url:    fmt.Sprintf("https://example.com/pfs/default/test/%v/big.txt", finishedCommit.Id),
			requestHeader: http.Header{
				"Range":    {"bytes=25000-25010"},
				"If-Match": {`"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`},
			},
			wantContent: "",
			wantCode:    http.StatusPreconditionFailed,
			wantHeader: http.Header{
				"Accept-Ranges": {"bytes"},
				"Cache-Control": {"private"},
				"Etag":          {`"ee45bb2661662b371f3100176c518243dbbe282116a05bb8f025a9c23b97ea02"`},
				"Last-Modified": {finishedAt.Format(http.TimeFormat)},
				"Vary":          {"authn-token"},
			},
		},
		{
			name:        "commit listing page, not implemented",
			method:      http.MethodGet,
			url:         "https://example.com/pfs/default/test",
			wantCode:    http.StatusNotFound,
			wantContent: "invalid URL; expecting /pfs/<project>/<repo>/<commit|branch>/<path...>, got /pfs/default/test",
		},
		{
			name:        "branch redirect",
			method:      http.MethodGet,
			url:         "https://example.com/pfs/default/test/master",
			wantCode:    http.StatusMovedPermanently,
			wantContent: "<a href=\"/pfs/default/test/master/\">Moved Permanently</a>.\n\n",
			wantHeader: http.Header{
				"Content-Type": {"text/html; charset=utf-8"},
				"Location":     {"/pfs/default/test/master/"},
				"Vary":         {"authn-token"},
			},
		},
		{
			name:        "directory listing of test@done",
			method:      http.MethodGet,
			url:         fmt.Sprintf("https://example.com/pfs/default/test/%v/", finishedCommit.Id),
			wantCode:    http.StatusOK,
			wantContent: "-\t26000\t/big.txt\nd\t13\t/sub/\n",
		},
		{
			name:   "html directory listing of test@done",
			method: http.MethodGet,
			requestHeader: http.Header{
				"Accept": {"text/html"},
			},
			url:         fmt.Sprintf("https://example.com/pfs/default/test/%v/", finishedCommit.Id),
			wantCode:    http.StatusOK,
			wantContent: `/(?s)href="big.txt">big.txt</a>.*href="sub">sub</a>/`,
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(test.method, test.url, nil)
			req = req.WithContext(pctx.TestContext(t))
			req.Header.Set(constants.ContextTokenKey, token.GetToken())
			for k, vs := range test.requestHeader {
				for _, v := range vs {
					req.Header.Add(k, v)
				}
			}
			dump, err := httputil.DumpRequest(req, false)
			if err != nil {
				t.Fatalf("dumping request failed: %v", err)
			}
			t.Logf("request: %s", dump)
			s.ServeHTTP(rec, req)
			dumpBody := true
			if req.Method == http.MethodHead {
				// The dumper is mad when the content-length doesn't match the body
				// size, but RFC9110 says "The server SHOULD send the same header
				// fields in response to a HEAD request as it would have sent if the
				// request method had been GET."  We do send content-length for GET
				// requests, so we also send them for HEAD requests.
				dumpBody = false
			}
			dump, err = httputil.DumpResponse(rec.Result(), dumpBody)
			if err != nil {
				t.Fatalf("dumping response failed: %v", err)
			}
			t.Logf("response: %s", dump)
			if got, want := rec.Code, test.wantCode; got != want {
				t.Errorf("response code:\n  got: %v\n want: %v", got, want)
			}
			if diff := cmp.Diff(test.wantContent, rec.Body.String(), cmputil.RegexpStrings()); diff != "" {
				t.Errorf("body (-want +got):\n%s", diff)
			}

			if want := test.wantHeader; want != nil {
				if diff := cmp.Diff(want, rec.Header()); diff != "" {
					t.Errorf("headers (-want +got):\n%s", diff)
				}
			}
		})
	}
}
