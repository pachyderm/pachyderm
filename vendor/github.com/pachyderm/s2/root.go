package s2

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// ListAllMyBucketsResult is an XML-encodable listing of repos as buckets
type ListAllMyBucketsResult struct {
	Owner   User     `xml:"Owner"`
	Buckets []Bucket `xml:"Buckets>Bucket"`
}

// Bucket is an XML-encodable repo, represented as an S3 bucket
type Bucket struct {
	Name         string    `xml:"Name"`
	CreationDate time.Time `xml:"CreationDate"`
}

type RootController interface {
	ListBuckets(r *http.Request, result *ListAllMyBucketsResult) error
}

type UnimplementedRootController struct{}

func (c UnimplementedRootController) ListBuckets(r *http.Request, result *ListAllMyBucketsResult) error {
	return NotImplementedError(r)
}

type rootHandler struct {
	controller RootController
	logger     *logrus.Entry
}

func (h *rootHandler) get(w http.ResponseWriter, r *http.Request) {
	result := &ListAllMyBucketsResult{}

	if err := h.controller.ListBuckets(r, result); err != nil {
		writeError(h.logger, w, r, err)
		return
	}

	writeXML(h.logger, w, r, http.StatusOK, result)
}
