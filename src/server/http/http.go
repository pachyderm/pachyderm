package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type router = *httprouter.Router

type server struct {
	router
	address string
}

func NewHTTPServer(address string) http.Handler {
	router := httprouter.New()
	result := server{
		router:  router,
		address: address,
	}
	router.GET("/v2/:RepoAndCommit/*Path", result.get)
	return result
}

func (s server) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
}
