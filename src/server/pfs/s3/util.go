package s3

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pachyderm/s2"
	"github.com/sirupsen/logrus"
)

// The S3 storage class that all PFS content will be reported to be stored in
const storageClass = "STANDARD"

// The S3 user associated with all PFS content
var defaultUser = s2.User{ID: "00000000000000000000000000000000", DisplayName: "pachyderm"}

func bucketArgs(r *http.Request, bucket string) (string, string, error) {
	parts := strings.SplitN(bucket, ".", 2)
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return "", "", s2.InvalidBucketNameError(r)
	}
	return parts[1], parts[0], nil
}

// TODO(ys): this is copy-pasta'd from s2, fix this
func writeError(logger *logrus.Entry, w http.ResponseWriter, r *http.Request, err *s2.Error) {
	vars := mux.Vars(r)
	requestID := vars["requestID"]

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-amz-id-2", requestID)
	w.Header().Set("x-amz-request-id", requestID)
	w.WriteHeader(err.HttpStatus)
	fmt.Fprint(w, xml.Header)

	if err := xml.NewEncoder(w).Encode(err); err != nil {
		// just log a message since a response has already been partially
		// written
		logger.Errorf("could not encode xml response: %v", err)
	}
}
