package spec

/* Generates spec document for our FUSE support
 *
 */

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"text/template"

	"errors"
	"fmt"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"path/filepath"
)

type Result int

const (
	UNDEFINED Result = 1 + iota
	UNSUPPORTED
	SUPPORTED
)

func (r Result) String() string {
	switch r {
	case UNDEFINED:
		return "undefined"
	case UNSUPPORTED:
		return "unsupported"
	case SUPPORTED:
		return "supported"
	}
	return ""
}

type Spec struct {
	Name    string
	Metric  string
	Results map[string]Result
	static  bool // Whether or not rows are defined by a static list or dynamically in tests
}

func New(name string, dataSet string) (*Spec, error) {
	tokens := strings.Split(filepath.Base(dataSet), ".")
	if len(tokens) < 2 {
		return nil, errors.New("Invalid filename - no name before the dot")
	}
	s := &Spec{
		Name:    name,
		Metric:  tokens[0],
		Results: make(map[string]Result),
		static:  true,
	}
	s.Load(dataSet)
	return s, nil
}

func (s *Spec) Load(dataSet string) error {
	data, err := ioutil.ReadFile(dataSet)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		s.Results[line] = UNDEFINED
	}

	return nil
}

// This function may be called several times (across tests). This is expected.
// We report the most conservative value in case you're debugging and tests
// dont pass and you're using the spec for reference.
//
// Basically - the spec results are an 'AND' not an 'OR'
func (s *Spec) NoError(t *testing.T, err error, result string) {
	state := UNSUPPORTED
	if err == nil {
		state = SUPPORTED
	}

	s.updateResult(t, result, state)
	require.NoError(t, err)
}

// This function may be called several times (across tests). This is expected.
// We report the most conservative value in case you're debugging and tests
// dont pass and you're using the spec for reference.
//
// Basically - the spec results are an 'AND' not an 'OR'
func (s *Spec) YesError(t *testing.T, err error, result string) {
	state := UNSUPPORTED
	if err == nil {
		state = SUPPORTED
	}
	s.updateResult(t, result, state)
	require.YesError(t, err)
}

func (s *Spec) updateResult(t *testing.T, resultName string, newResult Result) {
	oldResult, ok := s.Results[resultName]

	if !ok && s.static {
		// Missing this result row from the static spec. Err
		t.Errorf("Missing (%v) row from static spec (%v)\n", resultName, s.Name)
		return
	}

	// Leave this line. This will be helpful when using the tests to fix FUSE issues
	fmt.Printf("Checking spec for action: %v\n", newResult.String())

	switch oldResult {
	case UNSUPPORTED:
		// Cannot be assigned now that its in a unsupported state
	case UNDEFINED:
		// First time we've seen this check, assign the state
		s.Results[resultName] = newResult
	case SUPPORTED:
		// The previous state is successful, so assign the result of AND'ing
		// the previous state and the new state
		if oldResult == SUPPORTED && newResult == SUPPORTED {
			// Do nothing. Previous and current supported
		} else {
			// Demote to failure.
			s.Results[resultName] = UNSUPPORTED
		}
	}

}

func (s *Spec) GenerateReport(fileName string) error {
	t := template.New("spec.html")
	t, err := t.ParseFiles("spec/spec.html")
	if err != nil {
		return err
	}

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	err = t.Execute(w, s)
	if err != nil {
		return err
	}

	w.Flush()

	return nil
}
