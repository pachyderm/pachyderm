package s3

// TODO: the s2 library checks the type of the error to decide how to handle it,
// which doesn't work properly with wrapped errors

import (
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/s2"
)

func (c *controller) ListBuckets(r *http.Request) (*s2.ListBucketsResult, error) {
	defer log.Span(r.Context(), "ListBuckets")()

	pc := c.requestClient(r)
	result := s2.ListBucketsResult{
		Owner:   &defaultUser,
		Buckets: []*s2.Bucket{},
	}
	if err := c.driver.listBuckets(pc, r, &result.Buckets); err != nil {
		return nil, err
	}

	return &result, nil
}
