package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pachyderm/pachyderm/src/client/pfs"

	"golang.org/x/net/context"
)

const httpPort = 652

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
	fmt.Printf("STARTING HTTP SERVER\n")
	http.HandleFunc("/get-file", s.getFileHandler)
	return http.ListenAndServe(fmt.Sprintf(":%v", httpPort), nil)
}

func (s *HTTPServer) getFileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("rawquery: %v\n", r.URL.RawQuery)
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Printf("params: %v\n", params)
	pfsFile := &pfs.File{
		Commit: &pfs.Commit{
			ID: params.Get("file:commit:id"),
			Repo: &pfs.Repo{
				Name: params.Get("file:commit:repo:name"),
			},
		},
		Path: params.Get("file:path"),
	}
	offsetBytes, err := strconv.ParseInt(params.Get("offsetBytes"), 10, 64)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	sizeBytes, err := strconv.ParseInt(params.Get("sizeBytes"), 10, 64)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	file, err := s.driver.getFile(context.Background(), pfsFile, offsetBytes, sizeBytes)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", "attachment; filename=\"foo.txt\"")
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}
	io.Copy(&fw, file)
}
