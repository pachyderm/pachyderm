package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
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
}

func newHTTPServer(address string, etcdAddresses []string, etcdPrefix string, cacheBytes int64) (*HTTPServer, error) {
	d, err := newDriver(address, etcdAddresses, etcdPrefix, cacheBytes)
	if err != nil {
		return nil, err
	}
	router := httprouter.New()
	s := &HTTPServer{d, router}

	router.GET(fmt.Sprintf("/%v/pfs/repos/:repoName/commits/:commitID/files/*filePath", apiVersion), s.getFileHandler)
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

	file, err := s.driver.getFile(context.Background(), pfsFile, 0, 0)
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", fileName))
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}
	io.Copy(&fw, file)
}
