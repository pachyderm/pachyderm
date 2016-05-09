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
	FAILED
	SUCCEEDED
)

func (r Result) String() string {
	switch r {
	case UNDEFINED:
		return "undefined"
	case FAILED:
		return "failed"
	case SUCCEEDED:
		return "succeeded"
	}
	return ""
}

func (r Result) Bool() bool {
	if r == SUCCEEDED {
		return true
	}
	return false
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
	oldResult, ok := s.Results[result]

	if !ok && s.static {
		// Missing this result row from the static spec. Err
		t.Errorf("Missing (%v) row from static spec (%v)\n", result, s.Name)
		return
	}

	state := FAILED
	if err == nil {
		state = SUCCEEDED
	}

	// Leave this line. This will be helpful when using the tests to fix FUSE issues
	fmt.Printf("Spec check: %v\n", state.String())

	switch oldResult {
	case FAILED:
		// Cannot be assigned now that its in a failed state
	case UNDEFINED:
		// First time we've seen this check, assign the state
		s.Results[result] = state
	case SUCCEEDED:
		// The previous state is successful, so assign the result of AND'ing
		// the previous state and the new state
		if oldResult.Bool() && state.Bool() {
			// Do nothing. Previous and current succeeded
		} else {
			// Demote to failure.
			s.Results[result] = FAILED
		}
	}

	require.NoError(t, err)
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
