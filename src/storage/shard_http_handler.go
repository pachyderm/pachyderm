package storage

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
)

type shardHTTPHandler struct {
	*http.ServeMux
	Shard
}

func newShardHTTPHandler(shard Shard) *shardHTTPHandler {
	shardHTTPHandler := &shardHTTPHandler{
		http.NewServeMux(),
		shard,
	}
	shardHTTPHandler.HandleFunc("/branch", shardHTTPHandler.branch)
	shardHTTPHandler.HandleFunc("/commit", shardHTTPHandler.commit)
	shardHTTPHandler.HandleFunc("/file/", shardHTTPHandler.file)
	shardHTTPHandler.HandleFunc("/pipeline/", shardHTTPHandler.pipeline)
	shardHTTPHandler.HandleFunc("/pull", shardHTTPHandler.pull)
	shardHTTPHandler.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })
	return shardHTTPHandler
}

func resource(request *http.Request) string {
	url := strings.Split(request.URL.Path, "/")
	return path.Join(url[1:]...)
}

func (s *shardHTTPHandler) branch(writer http.ResponseWriter, request *http.Request) {
	name := resource(request)
	switch request.Method {
	case "GET":
		_, err := s.BranchGet(name)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	case "POST":
		branch, err := s.BranchCreate(name, commitParam(request))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(writer, "Created %s.", branch.Name)
	}
}

func (s *shardHTTPHandler) commit(writer http.ResponseWriter, request *http.Request) {
	name := resource(request)
	switch request.Method {
	case "GET":
		_, err := s.CommitGet(name)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	case "POST":
		commit, err := s.CommitCreate(name, branchParam(request))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(writer, "Created %s.", commit.Name)
	}
}

func (s *shardHTTPHandler) file(writer http.ResponseWriter, request *http.Request) {
	name := resource(request)
	switch request.Method {
	case "GET":
		file, err := s.FileGet(name, commitParam(request))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		http.ServeContent(writer, request, file.Name, file.ModTime, file.File)
	case "POST":
		if err := s.FileCreate(name, request.Body, branchParam(request)); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *shardHTTPHandler) pipeline(writer http.ResponseWriter, request *http.Request) {
	name := path.Join("pipeline", resource(request))
	switch request.Method {
	case "GET":
		file, err := s.FileGet(name, commitParam(request))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		http.ServeContent(writer, request, file.Name, file.ModTime, file.File)
	case "POST":
		if err := s.FileCreate(name, request.Body, branchParam(request)); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *shardHTTPHandler) pull(writer http.ResponseWriter, request *http.Request) {
	from := request.URL.Query().Get("from")
	mpw := multipart.NewWriter(writer)
	defer mpw.Close()
	cb := newMultipartPusher(mpw)
	writer.Header().Add("Boundary", mpw.Boundary())
	if err := s.Pull(from, cb); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}
