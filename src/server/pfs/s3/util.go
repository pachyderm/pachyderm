package s3

import (
	"encoding/xml"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pachyderm/pachyderm/src/client/pfs"
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

func bucketArgs(w http.ResponseWriter, r *http.Request, view map[string]*pfs.Commit) (string, string) {
	vars := mux.Vars(r)
	if view != nil {
		name := vars["name"]
		commit, ok := view[name]
		if !ok {
			// Returning empty repo and commit params, which are guaranteed to
			// be not found.
			return "", ""
		}
		return commit.Repo.Name, commit.ID
	}
	repo := vars["repo"]
	commit := vars["commit"]
	return repo, commit
}

func objectArgs(w http.ResponseWriter, r *http.Request, view map[string]*pfs.Commit) (string, string, string) {
	repo, commit := bucketArgs(w, r, view)
	vars := mux.Vars(r)
	file := vars["file"]
	return repo, commit, file
}

func requestLogger(r *http.Request) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"request-id": r.Header.Get("X-Request-ID"),
		"source":     "s3gateway",
	})
}
