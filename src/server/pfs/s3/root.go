package s3

import (
	"fmt"
	"net/http"

	"github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/s3server"
)

type rootController struct {
	pc     *client.APIClient
	logger *logrus.Entry
}

func (c rootController) List(r *http.Request, result *s3server.ListAllMyBucketsResult) *s3server.S3Error {
	result.Owner = defaultUser

	repos, err := c.pc.ListRepo()
	if err != nil {
		return s3server.InternalError(r, err)
	}

	for _, repo := range repos {
		t, err := types.TimestampFromProto(repo.Created)
		if err != nil {
			return s3server.InternalError(r, err)
		}

		for _, branch := range repo.Branches {
			result.Buckets = append(result.Buckets, s3server.Bucket{
				Name:         fmt.Sprintf("%s.%s", branch.Name, branch.Repo.Name),
				CreationDate: t,
			})
		}
	}

	return nil
}
