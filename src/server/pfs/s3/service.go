package s3

import (
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/s2"
)

func (c *controller) ListBuckets(r *http.Request) (*s2.ListBucketsResult, error) {
	c.logger.Debugf("ListBuckets")

	pc, err := c.requestClient(r)
	if err != nil {
		return nil, err
	}

	result := s2.ListBucketsResult{
		Owner:   &defaultUser,
		Buckets: []*s2.Bucket{},
	}
	if err = c.driver.listBuckets(pc, r, &result.Buckets); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return &result, nil
}
