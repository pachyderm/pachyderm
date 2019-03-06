package s3

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
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
	pc *client.APIClient
}

func newRootHandler(pc *client.APIClient) rootHandler {
	return rootHandler{pc: pc}
}

func (h rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	repos, err := h.pc.ListRepo()
	if err != nil {
		newInternalError(r, err).write(w)
		return
	}

	result := ListAllMyBucketsResult{
		Owner: defaultUser,
	}

	for _, repo := range repos {
		t, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			newInternalError(r, err).write(w)
			return
		}

		// NOTE: if `repo` has a `master` branch, this will not include
		// `repo-master` in the list of buckets, but rather just `repo`. This
		// is to avoid issues with conformance tests, which would try to
		// otherwise redundantly delete the `repo-master` bucket and fail.
		// TODO: figure out a more resilient way to manage this
		for _, branch := range repo.Branches {
			if branch.Name == "master" {
				// master branch can be addressed without it specified in the
				// bucket
				result.Buckets = append(result.Buckets, Bucket{
					Name:         branch.Repo.Name,
					CreationDate: t,
				})
			} else {
				result.Buckets = append(result.Buckets, Bucket{
					Name:         fmt.Sprintf("%s-%s", branch.Repo.Name, branch.Name),
					CreationDate: t,
				})
			}
		}
	}

	writeXML(w, http.StatusOK, &result)
}
