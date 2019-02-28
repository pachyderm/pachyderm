package s3

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

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

func writeBadRequest(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
}

func writeMaybeNotFound(w http.ResponseWriter, r *http.Request, err error) {
	if os.IsNotExist(err) || strings.Contains(err.Error(), "not found") {
		http.NotFound(w, r)
	} else {
		writeServerError(w, err)
	}
}

func writeServerError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
}

// writeXML serializes a struct to a response as XML
func writeXML(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(code)
	encoder := xml.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		// just log a message since a status code - and maybe part of
		logrus.Errorf("s3gateway: could not enocde xml response: %v", err)
	}
}

// intFormValue extracts an int value from a request's form values, ensuring
// it's within specified bounds. If `def` is non-nil, empty or unspecified
// form values default to it. Otherwise, an error is thrown. `r.ParseForm()`
// must be called before using this.
func intFormValue(r *http.Request, name string, min int, max int, def *int) (int, error) {
	s := r.FormValue(name)
	if s == "" {
		if def != nil {
			return *def, nil
		}
		return 0, fmt.Errorf("missing %s", name)
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value '%s': %s", name, s, err)
	}

	if i < min {
		return 0, fmt.Errorf("%s value %d cannot be less than %d", name, i, min)
	}
	if i > max {
		return 0, fmt.Errorf("%s value %d cannot be greater than %d", name, i, max)
	}

	return i, nil
}

// withBodyReader calls the provided callback with a reader for the HTTP
// request body. This also verifies the body against the `Content-MD5` header.
//
// The callback should return whether or not it succeeded. If it does not
// succeed, it is assumed that the callback wrote an appropriate failure
// response to the client.
//
// This function will return whether it succeeded and an error. If there is an
// error, it is because of a bad request. If this returns a failure but not an
// error, it implies that the callback returned a failure.
func withBodyReader(r *http.Request, f func(io.Reader) bool) (bool, error) {
	expectedHash := r.Header.Get("Content-MD5")

	if expectedHash != "" {
		expectedHashBytes, err := base64.StdEncoding.DecodeString(expectedHash)
		if err != nil {
			err = fmt.Errorf("could not decode `Content-MD5`, as it is not base64-encoded")
			return false, err
		}

		hasher := md5.New()
		reader := io.TeeReader(r.Body, hasher)

		succeeded := f(reader)
		if !succeeded {
			return false, nil
		}

		actualHash := hasher.Sum(nil)
		if !bytes.Equal(expectedHashBytes, actualHash) {
			err = fmt.Errorf("content checksums differ; expected=%x, actual=%x", expectedHash, actualHash)
			return false, err
		}

		return true, nil
	}

	return f(r.Body), nil
}

func objectArgs(r *http.Request) (string, string, string) {
	vars := mux.Vars(r)
	repo := vars["repo"]
	branch := vars["branch"]
	file := vars["file"]
	return repo, branch, file
}
