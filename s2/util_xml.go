package s2

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// WriteError serializes an error to a response as XML
func WriteError(logger *logrus.Entry, w http.ResponseWriter, r *http.Request, err error) {
	s3Err := newGenericError(r, err)
	writeXML(logger, w, r, s3Err.HTTPStatus, s3Err)
}

// writeXMLPrelude writes the HTTP headers and XML header to the response
func writeXMLPrelude(w http.ResponseWriter, r *http.Request, code int) {
	vars := mux.Vars(r)
	requestID := vars["requestID"]

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-amz-id-2", requestID)
	w.Header().Set("x-amz-request-id", requestID)
	w.WriteHeader(code)
	fmt.Fprint(w, xml.Header)
}

// writeXMLBody writes the marshaled XML payload of a value
func writeXMLBody(logger *logrus.Entry, w http.ResponseWriter, v interface{}) {
	encoder := xml.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		// just log a message since a response has already been partially
		// written
		logger.Errorf("could not encode xml response: %v", err)
	}
}

// writeXML writes HTTP headers, the XML header, and the XML payload to the
// response
func writeXML(logger *logrus.Entry, w http.ResponseWriter, r *http.Request, code int, v interface{}) {
	writeXMLPrelude(w, r, code)
	writeXMLBody(logger, w, v)
}

// readXMLBody reads an HTTP request body's bytes, and unmarshals it into
// `payload`.
func readXMLBody(r *http.Request, payload interface{}) error {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(bodyBytes, &payload)
	if err != nil {
		return MalformedXMLError(r)
	}
	return nil
}
