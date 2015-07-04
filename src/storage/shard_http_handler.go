package storage

import (
	"fmt"
	"net/http"
)

type shardHTTPHandler struct {
	*http.ServeMux
}

func newShardHTTPHandler(sIface Shard) *shardHTTPHandler {
	// TODO(pedge): remove when refactor done
	s, ok := sIface.(*shard)
	if !ok {
		panic("could not cast")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/branch", s.branchHandler)
	mux.HandleFunc("/commit", s.commitHandler)
	mux.HandleFunc("/file/", s.fileHandler)
	mux.HandleFunc("/pipeline/", s.pipelineHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })
	mux.HandleFunc("/pull", s.pullHandler)
	return &shardHTTPHandler{mux}
}
