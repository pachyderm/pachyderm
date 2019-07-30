package s3

import (
	"fmt"
	"net/http"

	"github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/s2"
)

type rootController struct {
	pc     *client.APIClient
	logger *logrus.Entry
}

func (c rootController) List(r *http.Request, result *s2.ListAllMyBucketsResult) error {
	result.Owner = defaultUser

	repos, err := c.pc.ListRepo()
	if err != nil {
		return s2.InternalError(r, err)
	}

	for _, repo := range repos {
		t, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			return s2.InternalError(r, err)
		}

		for _, branch := range repo.Branches {
			result.Buckets = append(result.Buckets, s2.Bucket{
				Name:         fmt.Sprintf("%s.%s", branch.Name, branch.Repo.Name),
				CreationDate: t,
			})
		}
	}

	return nil
}
