package s3

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/sirupsen/logrus"
)

type InputBucket struct {
	Repo     string
	CommitID string
	Name     string
}

type InputBuckets struct {
	values   []InputBucket
	reposMap map[string]*InputBucket
	namesMap map[string]*InputBucket
}

func NewInputBuckets(inputBuckets ...InputBucket) *InputBuckets {
	reposMap := map[string]*InputBucket{}
	namesMap := map[string]*InputBucket{}

	for _, ib := range inputBuckets {
		reposMap[ib.Repo] = &ib
		namesMap[ib.Name] = &ib
	}

	return &InputBuckets{
		values:   inputBuckets,
		reposMap: reposMap,
		namesMap: namesMap,
	}
}

type controller struct {
	pachdPort uint16

	logger *logrus.Entry

	// Name of the PFS repo holding multipart content
	repo string

	// the maximum number of allowed parts that can be associated with any
	// given file
	maxAllowedParts int

	// A list of buckets to serve, referencing specific commit IDs. If nil,
	// all PFS branches are served.
	inputBuckets *InputBuckets
}

func (c *controller) pachClient(authToken string) (*client.APIClient, error) {
	pc, err := client.NewFromAddress(fmt.Sprintf("localhost:%d", c.pachdPort))
	if err != nil {
		return nil, err
	}
	if authToken != "" {
		pc.SetAuthToken(authToken)
	}
	return pc, nil
}
