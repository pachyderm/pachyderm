package s3

import (
	"fmt"
	"net/http"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	"github.com/pachyderm/s2"
)

func (c *controller) ListBuckets(r *http.Request) (*s2.ListBucketsResult, error) {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return nil, err
	}
	repos, err := pc.ListRepo()
	if err != nil {
		return nil, err
	}

	result := s2.ListBucketsResult{
		Owner:   &defaultUser,
		Buckets: []s2.Bucket{},
	}

	for _, repo := range repos {
		t, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			return nil, err
		}

		if c.inputBuckets != nil {
			inputRepo := c.inputBuckets.reposMap[repo.Repo.Name]
			if inputRepo == nil {
				continue
			}

			result.Buckets = append(result.Buckets, s2.Bucket{
				Name:         inputRepo.Name,
				CreationDate: t,
			})
		} else {
			for _, branch := range repo.Branches {
				result.Buckets = append(result.Buckets, s2.Bucket{
					Name:         fmt.Sprintf("%s.%s", branch.Name, branch.Repo.Name),
					CreationDate: t,
				})
			}
		}
	}

	return &result, nil
}
