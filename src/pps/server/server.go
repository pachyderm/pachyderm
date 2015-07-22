package server

import "github.com/pachyderm/pachyderm/src/pps"

func NewAPIServer() pps.ApiServer {
	return newAPIServer()
}
