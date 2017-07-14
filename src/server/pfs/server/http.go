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

type HTTPServer struct {
	driver *driver
}

func newHTTPServer(address string, etcdAddresses []string, etcdPrefix string, cacheBytes int64) (*HTTPServer, error) {
	d, err := newDriver(address, etcdAddresses, etcdPrefix, cacheBytes)
	if err != nil {
		return nil, err
	}
	return &HTTPServer{d}, nil
}

func (s *HTTPServer) Start() error {
	router := httprouter.New()
	router.GET(fmt.Sprintf("/%v/pfs/repos/:repoName/commits/:commitID/files/*filePath", apiVersion), s.getFileHandler)
	// Serves a request like:
	// http://localhost:30652/v1/pfs/repos/foo/commits/b7a1923be56744f6a3f1525ec222dc3b/files/ttt.log
	return http.ListenAndServe(fmt.Sprintf(":%v", HTTPPort), router)
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

	offsetBytes := int64(0)
	sizeBytes := int64(0)
	file, err := s.driver.getFile(context.Background(), pfsFile, offsetBytes, sizeBytes)
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
