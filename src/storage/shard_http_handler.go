package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/src/btrfs"
	"github.com/pachyderm/pachyderm/src/common"
	"github.com/satori/go.uuid"
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
	shardHTTPHandler.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "%s\n", common.VersionString()) })
	return shardHTTPHandler
}

func (s *shardHTTPHandler) branch(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		branches, err := s.BranchList()
		if err != nil {
			httpError(writer, err)
			return
		}
		for _, branch := range branches {
			if err := json.NewEncoder(writer).Encode(branchMsg{Name: branch.Name, TStamp: branch.ModTime.Format("2006-01-02T15:04:05.999999-07:00")}); err != nil {
				httpError(writer, err)
				return
			}
		}
	case "POST":
		branch, err := s.BranchCreate(branchParam(request), commitParam(request))
		if err != nil {
			httpError(writer, err)
			return
		}
		fmt.Fprintln(writer, branch.Name)
	}
}

func (s *shardHTTPHandler) commit(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		commits, err := s.CommitList()
		if err != nil {
			httpError(writer, err)
			return
		}
		for _, commit := range commits {
			if err := json.NewEncoder(writer).Encode(commitMsg{Name: commit.Name, TStamp: commit.ModTime.Format("2006-01-02T15:04:05.999999-07:00")}); err != nil {
				httpError(writer, err)
				return
			}
		}
	case "POST":
		if request.ContentLength == 0 {
			values := request.URL.Query()
			if values.Get("commit") == "" {
				values.Add("commit", uuid.NewV4().String())
				request.URL.RawQuery = values.Encode()
			}
			commit, err := s.CommitCreate(commitParam(request), branchParam(request))
			if err != nil {
				httpError(writer, err)
				return
			}
			fmt.Fprintln(writer, commit.Name)
		} else {
			if err := s.Push(request.Body); err != nil {
				httpError(writer, err)
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
				httpError(writer, err)
				return
			}
			multiPartWriter := multipart.NewWriter(writer)
			defer multiPartWriter.Close()
			writer.Header().Add("Boundary", multiPartWriter.Boundary())
			for _, file := range files {
				fWriter, err := multiPartWriter.CreateFormFile(file.Name, file.Name)
				if err != nil {
					httpError(writer, err)
					return
				}
				if _, err := io.Copy(fWriter, file.File); err != nil {
					httpError(writer, err)
					return
				}
			}
		} else {
			file, err := s.FileGet(name, commitParam(request))
			if err != nil {
				httpError(writer, err)
				return
			}
			http.ServeContent(writer, request, file.Name, file.ModTime, file.File)
		}
	case "POST":
		if err := s.FileCreate(name, request.Body, branchParam(request)); err != nil {
			httpError(writer, err)
			return
		}
	}
}

func (s *shardHTTPHandler) pipeline(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		url := strings.Split(request.URL.Path, "/")
		pipelineName := url[2]
		fileName := path.Join(url[4:]...)
		if err := s.PipelineWait(pipelineName, commitParam(request)); err != nil {
			httpError(writer, err)
			return
		}
		if strings.ContainsAny(fileName, "*") {
			files, err := s.PipelineFileGetAll(pipelineName, fileName, commitParam(request), shardParam(request))
			if err != nil {
				httpError(writer, err)
				return
			}
			multiPartWriter := multipart.NewWriter(writer)
			defer multiPartWriter.Close()
			writer.Header().Add("Boundary", multiPartWriter.Boundary())
			for _, file := range files {
				fWriter, err := multiPartWriter.CreateFormFile(file.Name, file.Name)
				if err != nil {
					httpError(writer, err)
					return
				}
				if _, err := io.Copy(fWriter, file.File); err != nil {
					httpError(writer, err)
					return
				}
			}
		} else {
			file, err := s.PipelineFileGet(pipelineName, fileName, commitParam(request))
			if err != nil {
				httpError(writer, err)
				return
			}
			http.ServeContent(writer, request, file.Name, file.ModTime, file.File)
		}
	case "POST":
		name := resource(request)
		if err := s.PipelineCreate(name, request.Body, branchParam(request)); err != nil {
			httpError(writer, err)
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
		httpError(writer, err)
		return
	}
}

func resource(request *http.Request) string {
	url := strings.Split(request.URL.Path, "/")
	return path.Join(url[2:]...)
}

func httpError(writer http.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusInternalServerError)
}

func commitParam(r *http.Request) string {
	if p := r.URL.Query().Get("commit"); p != "" {
		return p
	}
	return "master"
}

func branchParam(r *http.Request) string {
	if p := r.URL.Query().Get("branch"); p != "" {
		return p
	}
	return "master"
}

func shardParam(r *http.Request) string {
	return r.URL.Query().Get("shard")
}

func hasBranch(r *http.Request) bool {
	return (r.URL.Query().Get("branch") == "")
}

func materializeParam(r *http.Request) string {
	if _, ok := r.URL.Query()["run"]; ok {
		return "true"
	}
	return "false"
}

func indexOf(haystack []string, needle string) int {
	for i, s := range haystack {
		if s == needle {
			return i
		}
	}
	return -1
}

func rawCat(w io.Writer, name string) error {
	f, err := btrfs.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		return err
	}
	return nil
}
