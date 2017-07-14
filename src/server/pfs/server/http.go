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
	fmt.Printf("STARTING HTTP SERVER\n")
	http.HandleFunc("/get-file", s.getFileHandler)
	return http.ListenAndServe(fmt.Sprintf(":%v", httpPort), nil)
}

func nextToken(tokens []string) (string, []string, error) {
	if len(tokens) == 0 {
		return "", nil, io.EOF
	}
	return tokens[0], tokens[1:len(tokens)], nil
}

func urlToFile(u URL) (*pfs.File, error) {
	reqPath := u.EscapedPath()
	fmt.Printf("path: %v\n", reqPath)
	tokens := strings.split(reqPath, "/")
	fmt.Printf("tokens: %v\n", tokens)
	// Example URL for v1
	// http://server.com/v1/pfs/repos/repoName/commits/commitID/files/path/to/the/real/file.txt

	// Parse API version / service
	next, tokens, err := nextToken(tokens)
	if err != nil {
		return nil, fmt.Errorf("invalid PFS api path %v", path)
	}
	if next != apiVersion {
		return nil, fmt.Errorf("unrecognized api version %v, expected %v", next, apiVersion)
	}
	next, tokens, err := nextToken(tokens)
	if err != nil {
		return nil, fmt.Errorf("invalid PFS api path %v", path)
	}
	if next != "pfs" {
		return nil, fmt.Errorf("invalid PFS api path %v expecting 'pfs' service", path)
	}

	// parse repo name
	next, tokens, err := nextToken(tokens)
	if err != nil {
		return nil, fmt.Errorf("invalid PFS api path %v", path)
	}
	if next != "repos" {
		return nil, fmt.Errorf("invalid PFS api path %v expecting 'repos' to be provided", path)
	}
	next, tokens, err := nextToken(tokens)
	if err != nil {
		return nil, fmt.Errorf("invalid PFS api path %v", path)
	}
	repoName := next

	// parse commit ID
	next, tokens, err := nextToken(tokens)
	if err != nil {
		return nil, fmt.Errorf("invalid PFS api path %v", path)
	}
	if next != "commits" {
		return nil, fmt.Errorf("invalid PFS api path %v expecting 'commits' to be provided", path)
	}
	next, tokens, err := nextToken(tokens)
	if err != nil {
		return nil, fmt.Errorf("invalid PFS api path %v", path)
	}
	commitID := next

	// parse file path
	next, tokens, err := nextToken(tokens)
	if err != nil {
		return nil, fmt.Errorf("invalid PFS api path %v", path)
	}
	if next != "files" {
		return nil, fmt.Errorf("invalid PFS api path %v expecting 'files' to be provided", path)
	}
	var filePaths []string

	for next, tokens, err := nextToken(tokens); err != io.EOF; next, tokens, err := nextToken(tokens) {
		if err != nil {
			return nil, fmt.Errorf("invalid PFS api path %v", path)
		}
		filePaths = append(filePaths, next)
	}

	return &pfs.File{
		Commit: &pfs.Commit{
			ID: commitID,
			Repo: &pfs.Repo{
				Name: repoName,
			},
		},
		Path: strings.Join(filePaths, "/"),
	}, nil
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
