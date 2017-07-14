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

func parseFieldFromURL(fieldName string, tokens []string, rawPath string) (value string, tokens []string, err error) {
	value, tokens, err = nextToken(tokens)
	if err != nil {
		return "", nil, fmt.Errorf("invalid PFS api path %v", rawPath)
	}
	if value != fieldName {
		return "", nil, fmt.Errorf("invalid PFS api path %v expecting '%v' to be provided", rawPath, fieldName)
	}
	value, tokens, err = nextToken(tokens)
	if err != nil {
		return "", nil, fmt.Errorf("invalid PFS api path %v", rawPath)
	}
	return value, tokens, nil
}

func urlToFile(u URL) (file *pfs.File, fileName string, err error) {
	reqPath := u.EscapedPath()
	fmt.Printf("path: %v\n", reqPath)
	tokens := strings.split(reqPath, "/")
	fmt.Printf("tokens: %v\n", tokens)

	// Example URL for v1
	// http://server.com/v1/pfs/repos/repoName/commits/commitID/files/path/to/the/real/file.txt

	version, tokens, err := parseFieldFromURL("v1", tokens, reqPath)
	next, tokens, err := nextToken(tokens)
	if err != nil {
		return nil, err
	}
	service, tokens, err := parseFieldFromURL("pfs", tokens, reqPath)
	if err != nil {
		return nil, err
	}
	repoName, tokens, err := parseFieldFromURL("repos", tokens, reqPath)
	if err != nil {
		return nil, err
	}
	commitID, tokens, err := parseFieldFromURL("commits", tokens, reqPath)
	if err != nil {
		return nil, err
	}

	// parse file path
	var filePaths []string
	next, tokens, err := parseFieldFromURL("files", tokens, reqPath)
	if err != nil {
		return nil, err
	}
	filePaths = append(filePaths, next)
	for ; err != io.EOF; next, tokens, err := nextToken(tokens) {
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
	}, filePaths[len(filePaths)-1], nil
}

func (s *HTTPServer) getFileHandler(w http.ResponseWriter, r *http.Request) {
	pfsFile, fileName, err := urlToFile(r.URL)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	offsetBytes := 0
	sizeBytes := 0
	file, err := s.driver.getFile(context.Background(), pfsFile, offsetBytes, sizeBytes)
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add(fmt.Sprintf("Content-Disposition", "attachment; filename=\"%v\"", fileName))
	fw := flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}
	io.Copy(&fw, file)
}
