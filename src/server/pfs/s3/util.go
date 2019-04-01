package s3

import (
	"encoding/xml"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// The S3 storage class that all PFS content will be reported to be stored in
const storageClass = "STANDARD"

// The S3 user associated with all PFS content
var defaultUser = User{ID: "00000000000000000000000000000000", DisplayName: "pachyderm"}

// User is an XML-encodable representation of an S3 user
type User struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

// writeXML serializes a struct to a response as XML
func writeXML(w http.ResponseWriter, r *http.Request, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(code)
	encoder := xml.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		// just log a message since a response has already been partially
		// written
		requestLogger(r).Errorf("could not encode xml response: %v", err)
	}
}

func bucketArgs(w http.ResponseWriter, r *http.Request) (string, string) {
	vars := mux.Vars(r)
	repo := vars["repo"]
	branch := vars["branch"]
	return repo, branch
}

func objectArgs(w http.ResponseWriter, r *http.Request) (string, string, string) {
	repo, branch := bucketArgs(w, r)
	vars := mux.Vars(r)
	file := vars["file"]
	return repo, branch, file
}

func requestLogger(r *http.Request) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"request-id": r.Header.Get("X-Request-ID"),
		"source": "s3gateway",
	})
}
