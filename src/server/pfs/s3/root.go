package s3

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// ListAllMyBucketsResult is an XML-encodable listing of repos as buckets
type ListAllMyBucketsResult struct {
	Owner   User     `xml:"Owner"`
	Buckets []Bucket `xml:"Buckets>Bucket"`
}

// Bucket is an XML-encodable repo, represented as an S3 bucket
type Bucket struct {
	Name         string    `xml:"Name"`
	CreationDate time.Time `xml:"CreationDate"`
}

type rootHandler struct {
	pc   *client.APIClient
	view map[string]*pfs.Commit
}

func newRootHandler(pc *client.APIClient, view map[string]*pfs.Commit) rootHandler {
	return rootHandler{pc: pc, view: view}
}

func (h rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result := ListAllMyBucketsResult{
		Owner: defaultUser,
	}

	if h.view != nil {
		for name, commit := range h.view {
			ri, err := h.pc.InspectRepo(commit.Repo.Name)
			if err != nil {
				internalError(w, r, err)
				return
			}
			t, err := types.TimestampFromProto(ri.Created)
			if err != nil {
				internalError(w, r, err)
				return
			}
			result.Buckets = append(result.Buckets, Bucket{
				Name:         name,
				CreationDate: t,
			})
		}
	} else {
		repos, err := h.pc.ListRepo()
		if err != nil {
			internalError(w, r, err)
			return
		}
		for _, repo := range repos {
			t, err := types.TimestampFromProto(repo.Created)
			if err != nil {
				internalError(w, r, err)
				return
			}

			for _, branch := range repo.Branches {
				result.Buckets = append(result.Buckets, Bucket{
					Name:         fmt.Sprintf("%s.%s", branch.Name, branch.Repo.Name),
					CreationDate: t,
				})
			}
		}
	}

	writeXML(w, r, http.StatusOK, &result)
}
