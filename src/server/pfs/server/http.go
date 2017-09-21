package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// HTTPPort specifies the port the server will listen on
const HTTPPort = 652
const apiVersion = "v1"

type flushWriter struct {
	f http.Flusher
	w io.Writer
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if fw.f != nil {
		fw.f.Flush()
	}
	return
}

// HTTPServer serves GetFile requests over HTTP
// e.g. http://localhost:30652/v1/pfs/repos/foo/commits/b7a1923be56744f6a3f1525ec222dc3b/files/ttt.log
type HTTPServer struct {
	driver *driver
	*httprouter.Router
	loginPath string
}

func newHTTPServer(address string, etcdAddresses []string, etcdPrefix string, cacheSize int64) (*HTTPServer, error) {
	d, err := newDriver(address, etcdAddresses, etcdPrefix, cacheSize)
	if err != nil {
		return nil, err
	}
	router := httprouter.New()
	s := &HTTPServer{
		d,
		router,
		fmt.Sprintf("/%v/auth/login", apiVersion),
	}

	router.GET(fmt.Sprintf("/%v/pfs/repos/:repoName/commits/:commitID/files/*filePath", apiVersion), s.getFileHandler)
	router.POST(s.loginPath, s.authLoginHandler)
	router.POST(fmt.Sprintf("/%v/auth/logout", apiVersion), s.authLogoutHandler)
	// Debug method (to check login cookies):
	router.GET(s.loginPath, s.loginForm)
	router.NotFound = &notFoundRouter{}
	return s, nil
}

func (s *HTTPServer) getFileHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pfsFile := &pfs.File{
		Commit: &pfs.Commit{
			ID: ps.ByName("commitID"),
			Repo: &pfs.Repo{
				Name: ps.ByName("repoName"),
			},
		},
		Path: ps.ByName("filePath"),
	}
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
	// Since we can't seek, open a separate reader to sniff mimetype
	mimeReader, err := s.driver.getFile(ctx, pfsFile, 0, 0)
	if err != nil {
		panic(err)
	}
	buffer := make([]byte, 512)
	_, err = mimeReader.Read(buffer)
	if err != nil && err != io.EOF {
		panic(err)
	}
	contentType := http.DetectContentType(buffer)

	file, err := s.driver.getFile(ctx, pfsFile, 0, 0)
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", contentType)
	downloadValues := r.URL.Query()["download"]
	if len(downloadValues) == 1 && downloadValues[0] == "true" {
		w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", fileName))
	}
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}
	io.Copy(&fw, file)
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

type notFoundRouter struct {
}

func (s notFoundRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	b := strings.NewReader("route not found")
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}
	io.Copy(&fw, b)
}
func (s *HTTPServer) loginForm(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	content := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body>
<form action="%v" method="post">
<input name="Token">
<input type="submit">
</form>
</body>
</html>
	`, s.loginPath)
	b := strings.NewReader(content)
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}
	io.Copy(&fw, b)
}
