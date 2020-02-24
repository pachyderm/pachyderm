package s3

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pachyderm/s2"
)

func (c *controller) ListBuckets(r *http.Request) (*s2.ListBucketsResult, error) {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return nil, err
	}

	result := s2.ListBucketsResult{
		Owner:   &defaultUser,
		Buckets: []s2.Bucket{},
	}
	if err = c.driver.ListBuckets(pc, r, &result.Buckets); err != nil {
		return nil, err
	}
	
	return &result, nil
}
