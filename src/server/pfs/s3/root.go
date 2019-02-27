package s3

import (
	"net/http"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/gogo/protobuf/types"
)

type ListAllMyBucketsResult struct {
	Owner User `xml:"Owner"`
	Buckets []Bucket `xml:"Buckets>Bucket"`
}

type Bucket struct {
	Name string `xml:"Name"`
	CreationDate time.Time `xml:"CreationDate"`
}

type rootHandler struct {
	pc           *client.APIClient
}

func newRootHandler(pc *client.APIClient) rootHandler {
	return rootHandler{pc: pc}
}

func (h rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	repos, err := h.pc.ListRepo()
	if err != nil {
		writeServerError(w, err)
		return
	}

	result := ListAllMyBucketsResult{
		Owner: defaultUser,
	}

	for _, repo := range repos {
		t, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			writeServerError(w, err)
			return
		}

		result.Buckets = append(result.Buckets, Bucket{
			Name: repo.Repo.Name,
			CreationDate: t,
		})
	}

	writeXML(w, http.StatusOK, &result)
}
