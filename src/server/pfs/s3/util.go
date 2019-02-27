package s3

import (
	"fmt"
	"net/http"
	"encoding/xml"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const storageClass = "STANDARD"

var defaultUser = User { ID: "00000000000000000000000000000000", DisplayName: "pachyderm" }

type User struct {
	ID string `xml:"ID"`
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

func writeXML(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(code)
	encoder := xml.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
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
