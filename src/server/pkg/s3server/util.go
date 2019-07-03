package s3server

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// writeXML serializes a struct to a response as XML
func writeXML(logger *logrus.Entry, w http.ResponseWriter, r *http.Request, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(code)
	fmt.Fprintf(w, xml.Header)
	encoder := xml.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		// just log a message since a response has already been partially
		// written
		logger.Errorf("could not encode xml response: %v", err)
	}
}

func NotImplementedEndpoint(logger *logrus.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		NotImplementedError(r).Write(logger, w)
	}
}
