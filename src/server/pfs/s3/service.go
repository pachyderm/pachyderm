package s3

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/s2"
)

type serviceController struct {
	pc     *client.APIClient
	logger *logrus.Entry
}

func newServiceController(pc *client.APIClient, logger *logrus.Entry) *serviceController {
	c := serviceController{
		pc:     pc,
		logger: logger,
	}

	return &c
}

func (c *serviceController) ListBuckets(r *http.Request) (owner *s2.User, buckets []s2.Bucket, err error) {
	repos, err := c.pc.ListRepo()
	if err != nil {
		return
	}

	for _, repo := range repos {
		var t time.Time
		t, err = types.TimestampFromProto(repo.Created)
		if err != nil {
			return
		}

		for _, branch := range repo.Branches {
			buckets = append(buckets, s2.Bucket{
				Name:         fmt.Sprintf("%s.%s", branch.Name, branch.Repo.Name),
				CreationDate: t,
			})
		}
	}

	owner = &defaultUser
	return
}
