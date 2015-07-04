package storage

import (
	"fmt"
	"net/http"
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

}

func (s *shardHTTPHandler) commit(writer http.ResponseWriter, request *http.Request) {

}

func (s *shardHTTPHandler) file(writer http.ResponseWriter, request *http.Request) {

}

func (s *shardHTTPHandler) pipeline(writer http.ResponseWriter, request *http.Request) {

}

func (s *shardHTTPHandler) pull(writer http.ResponseWriter, request *http.Request) {

}
