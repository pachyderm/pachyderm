package s2

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// writeError serializes an error to a response as XML
func writeError(logger *logrus.Entry, w http.ResponseWriter, r *http.Request, err error) {
	switch e := err.(type) {
	case *Error:
		writeXML(logger, w, r, e.HttpStatus, e)
	default:
		s3Err := InternalError(r, e)
		writeXML(logger, w, r, s3Err.HttpStatus, s3Err)
	}
}

func writeXMLPrelude(w http.ResponseWriter, r *http.Request, code int) {
	vars := mux.Vars(r)
	requestID := vars["requestID"]

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-amz-id-2", requestID)
	w.Header().Set("x-amz-request-id", requestID)
	w.WriteHeader(code)
	fmt.Fprint(w, xml.Header)
}

func writeXMLBody(logger *logrus.Entry, w http.ResponseWriter, v interface{}) {
	encoder := xml.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		// just log a message since a response has already been partially
		// written
		logger.Errorf("could not encode xml response: %v", err)
	}
}

func writeXML(logger *logrus.Entry, w http.ResponseWriter, r *http.Request, code int, v interface{}) {
	writeXMLPrelude(w, r, code)
	writeXMLBody(logger, w, v)
}

func NotImplementedEndpoint(logger *logrus.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(logger, w, r, NotImplementedError(r))
	}
}

func withBodyReader(r *http.Request, f func(reader io.Reader) error) (bool, error) {
	expectedHash, ok := r.Header["Content-Md5"]
	var expectedHashBytes []uint8
	var err error
	if ok && len(expectedHash) == 1 {
		expectedHashBytes, err = base64.StdEncoding.DecodeString(expectedHash[0])
		if err != nil || len(expectedHashBytes) != 16 {
			return false, InvalidDigestError(r)
		}
	}

	hasher := md5.New()
	reader := io.TeeReader(r.Body, hasher)
	if err = f(reader); err != nil {
		return false, err
	}

	actualHashBytes := hasher.Sum(nil)
	if expectedHashBytes != nil && !bytes.Equal(expectedHashBytes, actualHashBytes) {
		return true, BadDigestError(r)
	}

	return false, nil
}

// intFormValue extracts an int value from a request's form values, ensuring
// it's within specified bounds. If the value is unspecified, `def` is
// returned. If the value is not an int, or not with the specified bounds, an
// error is returned.
func intFormValue(r *http.Request, name string, min int, max int, def int) (int, error) {
	s := r.FormValue(name)
	if s == "" {
		return def, nil
	}

	i, err := strconv.Atoi(s)
	if err != nil || i < min || i > max {
		return 0, InvalidArgument(r)
	}

	return i, nil
}

func stripETagQuotes(s string) string {
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return strings.Trim(s, "\"")
	}
	return s
}

func addETagQuotes(s string) string {
	if !strings.HasPrefix(s, "\"") {
		return fmt.Sprintf("\"%s\"", s)
	}
	return s
}
