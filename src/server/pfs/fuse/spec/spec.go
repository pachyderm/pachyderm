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

	"fmt"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"text/template"
)

type Result int

const (
	UNDEFINED Result = 1 + iota
	FAILED
	SUCCEEDED
)

type Spec struct {
	Name    string
	Results map[string]Result
}

func New(name string) *Spec {
	return &Spec{
		Name:    name,
		Results: make(map[string]Result),
	}
}

func (s *Spec) Load(dataSet string) error {
	data, err := ioutil.ReadFile(dataSet)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
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

func (s *Spec) fileName() string {
	normalized := strings.Replace(s.Name, " ", "-", -1)
	return fmt.Sprintf("%v-report.html", normalized)
}

func (s *Spec) GenerateReport() error {
	t := template.New("report")
	t, err := t.ParseFiles("spec.html")
	if err != nil {
		return err
	}

	f, err := os.Create(s.fileName())
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
