package s3

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	"github.com/pachyderm/s2"
	"github.com/sirupsen/logrus"
)

type serviceController struct {
	logger *logrus.Entry
}

func (c serviceController) ListBuckets(r *http.Request) (owner *s2.User, buckets []s2.Bucket, err error) {
	vars := mux.Vars(r)
	pc, err := pachClient(vars["authAccessKey"])
	if err != nil {
		return
	}
	repos, err := pc.ListRepo()
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
