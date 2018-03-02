package http

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"

	"github.com/gogo/protobuf/types"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// HTTPPort specifies the port the server will listen on
const HTTPPort = 652
const apiVersion = "v1"

func versionPath(p string) string {
	return path.Join("/", apiVersion, p)
}

var (
	getFilePath = versionPath("pfs/repos/:repoName/commits/:commitID/files/*filePath")
	loginPath   = versionPath("auth/login")
	logoutPath  = versionPath("auth/logout")
)

type router = *httprouter.Router

// HTTPServer serves GetFile requests over HTTP
// e.g. http://localhost:30652/v1/pfs/repos/foo/commits/b7a1923be56744f6a3f1525ec222dc3b/files/ttt.log
type HTTPServer struct {
	router
	address        string
	pachClient     *client.APIClient
	pachClientOnce sync.Once
}

// NewHTTPServer returns a Pachyderm HTTP server.
func NewHTTPServer(address string) (http.Handler, error) {
	router := httprouter.New()
	s := &HTTPServer{
		router:  router,
		address: address,
	}

	router.GET(getFilePath, s.getFileHandler)
	router.POST(loginPath, s.authLoginHandler)
	router.POST(logoutPath, s.authLogoutHandler)
	// Debug method (to check login cookies):
	router.GET(loginPath, s.loginForm)
	router.NotFound = http.HandlerFunc(notFound)
	return s, nil
}

func (s *HTTPServer) getFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filePaths := strings.Split(ps.ByName("filePath"), "/")
	fileName := filePaths[len(filePaths)-1]
	ctx := context.Background()
	for _, cookie := range r.Cookies() {
		if cookie.Name == auth.ContextTokenKey {
			ctx = metadata.NewIncomingContext(
				ctx,
				metadata.Pairs(auth.ContextTokenKey, cookie.Value),
			)
		}
	}
	downloadValues := r.URL.Query()["download"]
	if len(downloadValues) == 1 && downloadValues[0] == "true" {
		w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", fileName))
	}
	c := s.getPachClient()
	commitInfo, err := c.InspectCommit(ps.ByName("repoName"), ps.ByName("commitID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	content, err := c.GetFileReadSeeker(ps.ByName("repoName"), ps.ByName("commitID"), ps.ByName("filePath"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	modtime, err := types.TimestampFromProto(commitInfo.Finished)
	http.ServeContent(w, r, fileName, modtime, content)
	return
}

type loginRequestPayload struct {
	Token string
}

func (s *HTTPServer) authLoginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token := r.FormValue("Token")
	if token == "" {
		// Return 500
		http.Error(w, "empty token provided", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Set-Cookie", fmt.Sprintf("%v=%v;path=/", auth.ContextTokenKey,
		token))
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
}

func (s *HTTPServer) authLogoutHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Set-Cookie", fmt.Sprintf("%v=;path=/", auth.ContextTokenKey))
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	http.Error(w, "route not found", http.StatusNotFound)
}
func (s *HTTPServer) loginForm(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html>
<html>
<body>
<form action="%v" method="post">
<input name="Token">
<input type="submit">
</form>
</body>
</html>
`, loginPath))); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *HTTPServer) getPachClient() *client.APIClient {
	s.pachClientOnce.Do(func() {
		var err error
		s.pachClient, err = client.NewFromAddress(s.address)
		if err != nil {
			panic(fmt.Sprintf("http server failed to initialize pach client: %v", err))
		}
	})
	return s.pachClient
}
