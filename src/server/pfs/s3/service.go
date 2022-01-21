//nolint:wrapcheck
// TODO: the s2 library checks the type of the error to decide how to handle it,
// which doesn't work properly with wrapped errors
package s3

import (
	"net/http"

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
		return nil, err
	}

	return &result, nil
}
