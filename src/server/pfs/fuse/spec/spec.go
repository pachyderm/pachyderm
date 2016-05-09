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

type Spec struct {
	Name    string
	Metric  string
	Results map[string]Result
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

func (s *Spec) NoError(t *testing.T, err error, result string) {
	state := FAILED
	if err == nil {
		state = SUCCEEDED
	}
	s.Results[result] = state
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
