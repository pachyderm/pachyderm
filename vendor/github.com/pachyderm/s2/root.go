package s2

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Bucket is an XML marshalable representation of a bucket
type Bucket struct {
	Name         string    `xml:"Name"`
	CreationDate time.Time `xml:"CreationDate"`
}

// ServiceController is an interface defining service-level functionality
type ServiceController interface {
	// ListBuckets lists all buckets
	ListBuckets(r *http.Request) (owner *User, buckets []Bucket, err error)
}

// unimplementedServiceController defines a controller that returns
// `NotImplementedError` for all functionality
type unimplementedServiceController struct{}

func (c unimplementedServiceController) ListBuckets(r *http.Request) (owner *User, buckets []Bucket, err error) {
	return nil, nil, NotImplementedError(r)
}

type serviceHandler struct {
	controller ServiceController
	logger     *logrus.Entry
}

func (h *serviceHandler) get(w http.ResponseWriter, r *http.Request) {
	owner, buckets, err := h.controller.ListBuckets(r)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	writeXML(h.logger, w, r, http.StatusOK, struct {
		XMLName xml.Name `xml:"ListAllMyBucketsResult"`
		Owner   *User    `xml:"Owner"`
		Buckets []Bucket `xml:"Buckets>Bucket"`
	}{
		Owner:   owner,
		Buckets: buckets,
	})
}
