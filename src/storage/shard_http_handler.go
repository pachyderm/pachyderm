package storage

import (
	"encoding/json"
	"fmt"
	"io"
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

func (s *shardHTTPHandler) branch(writer http.ResponseWriter, request *http.Request) {
	name := resource(request)
	switch request.Method {
	case "GET":
		branches, err := s.BranchList()
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		encoder := json.NewEncoder(writer)
		for _, branch := range branches {
			err = encoder.Encode(branchMsg{Name: branch.Name, TStamp: branch.ModTime.Format("2006-01-02T15:04:05.999999-07:00")})
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	case "POST":
		branch, err := s.BranchCreate(name, commitParam(request))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(writer, "Created %s.", branch.Name)
	}
}

func (s *shardHTTPHandler) commit(writer http.ResponseWriter, request *http.Request) {
	name := resource(request)
	switch request.Method {
	case "GET":
		commits, err := s.BranchList()
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		encoder := json.NewEncoder(writer)
		for _, commit := range commits {
			err = encoder.Encode(branchMsg{Name: commit.Name, TStamp: commit.ModTime.Format("2006-01-02T15:04:05.999999-07:00")})
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	case "POST":
		if request.ContentLength == 0 {
			commit, err := s.CommitCreate(name, branchParam(request))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(writer, "Created %s.", commit.Name)
		} else {
			if err := s.Push(request.Body); err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func (s *shardHTTPHandler) file(writer http.ResponseWriter, request *http.Request) {
	name := resource(request)
	switch request.Method {
	case "GET":
		if strings.ContainsAny(name, "*") {
			files, err := s.FileGetAll(name, commitParam(request))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			multiPartWriter := multipart.NewWriter(writer)
			defer multiPartWriter.Close()
			writer.Header().Add("Boundary", multiPartWriter.Boundary())
			for _, file := range files {
				fWriter, err := multiPartWriter.CreateFormFile(file.Name, file.Name)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}
				if _, err := io.Copy(fWriter, file.File); err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else {
			file, err := s.FileGet(name, commitParam(request))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			http.ServeContent(writer, request, file.Name, file.ModTime, file.File)
		}
	case "POST":
		if err := s.FileCreate(name, request.Body, branchParam(request)); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *shardHTTPHandler) pipeline(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		url := strings.Split(request.URL.Path, "/")
		pipelineName := url[1]
		fileName := path.Join(url[3:]...)
		if err := s.PipelineWait(pipelineName, commitParam(request)); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if strings.ContainsAny(fileName, "*") {
			files, err := s.FileGetAll(fileName, commitParam(request))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			multiPartWriter := multipart.NewWriter(writer)
			defer multiPartWriter.Close()
			writer.Header().Add("Boundary", multiPartWriter.Boundary())
			for _, file := range files {
				fWriter, err := multiPartWriter.CreateFormFile(file.Name, file.Name)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}
				if _, err := io.Copy(fWriter, file.File); err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else {
			file, err := s.PipelineFileGet(pipelineName, fileName, commitParam(request))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			http.ServeContent(writer, request, file.Name, file.ModTime, file.File)
		}
	case "POST":
		name := resource(request)
		if err := s.PipelineCreate(name, request.Body, branchParam(request)); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
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

func resource(request *http.Request) string {
	url := strings.Split(request.URL.Path, "/")
	return path.Join(url[1:]...)
}
