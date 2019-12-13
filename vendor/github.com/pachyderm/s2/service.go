package s2

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Bucket is an XML marshallable representation of a bucket
type Bucket struct {
	// Name is the bucket name
	Name string `xml:"Name"`
	// CreationDate is when the bucket was created
	CreationDate time.Time `xml:"CreationDate"`
}

// ListBucketsResult is a response from a ListBucket call
type ListBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	// Owner is the owner of the buckets
	Owner *User `xml:"Owner"`
	// Buckets are a list of buckets under the given owner
	Buckets []Bucket `xml:"Buckets>Bucket"`
}

// ServiceController is an interface defining service-level functionality
type ServiceController interface {
	// ListBuckets lists all buckets
	ListBuckets(r *http.Request) (*ListBucketsResult, error)
}

// unimplementedServiceController defines a controller that returns
// `NotImplementedError` for all functionality
type unimplementedServiceController struct{}

func (c unimplementedServiceController) ListBuckets(r *http.Request) (*ListBucketsResult, error) {
	return nil, NotImplementedError(r)
}

type serviceHandler struct {
	controller ServiceController
	logger     *logrus.Entry
}

func (h *serviceHandler) get(w http.ResponseWriter, r *http.Request) {
	result, err := h.controller.ListBuckets(r)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	writeXML(h.logger, w, r, http.StatusOK, result)
}
