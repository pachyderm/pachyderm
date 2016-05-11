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
	"runtime"
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
	Results map[string]Result // key = action name, value = result (if its supported)
	static  bool              // Whether or not rows are defined by a static list or dynamically in tests
	*Coverage
}

type Coverage struct {
	UndefinedCount        int
	UndefinedPercentage   float64
	UnsupportedCount      int
	UnsupportedPercentage float64
	SupportedCount        int
	SupportedPercentage   float64
	TotalCount            int
}

func NewCoverage(undefined int, unsupported int, supported int) *Coverage {
	total := undefined + unsupported + supported
	return &Coverage{
		UndefinedCount:        undefined,
		UndefinedPercentage:   float64(undefined) / float64(total),
		UnsupportedCount:      unsupported,
		UnsupportedPercentage: float64(unsupported) / float64(total),
		SupportedCount:        supported,
		SupportedPercentage:   float64(supported) / float64(total),
		TotalCount:            total,
	}
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

// NoError  may be called several times (across tests). This is expected.
// We report the most conservative value in case you're debugging, tests
// don't pass, and you're using the spec for reference.
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

// YesError may be called several times (across tests). This is expected.
// We report the most conservative value in case you're debugging, tests
// don't pass, and you're using the spec for reference.
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

func (s *Spec) CalculateCoverage() {
	undefined := 0
	unsupported := 0
	supported := 0
	for _, result := range s.Results {
		switch result {
		case UNDEFINED:
			undefined += 1
		case UNSUPPORTED:
			unsupported += 1
		case SUPPORTED:
			supported += 1
		}
	}

	s.Coverage = NewCoverage(undefined, unsupported, supported)
}

func (s *Spec) GenerateReport(fileName string) error {
	t, err := template.ParseFiles("spec/spec.html")
	if err != nil {
		return err
	}

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	s.CalculateCoverage()

	err = t.ExecuteTemplate(w, "spec", s)
	if err != nil {
		return err
	}

	w.Flush()

	return nil
}

type CombinedSpec struct {
	Metric    string
	SpecNames []string
	Results   map[string][]Result // key = action, value = slice of result values
	*Coverage
}

func NewCombinedSpec(specs []Spec) *CombinedSpec {
	cs := &CombinedSpec{
		Metric:  specs[0].Metric,
		Results: make(map[string][]Result),
	}

	for _, spec := range specs {
		cs.SpecNames = append(cs.SpecNames, spec.Name)
		for name, result := range spec.Results {
			cs.Results[name] = append(cs.Results[name], result)
		}
	}

	return cs
}

func (cs *CombinedSpec) CalculateCoverage() {
	undefined := 0
	unsupported := 0
	supported := 0

	for _, results := range cs.Results {
		for _, result := range results {
			switch result {
			case UNDEFINED:
				undefined += 1
			case UNSUPPORTED:
				unsupported += 1
			case SUPPORTED:
				supported += 1
			}
		}
	}

	cs.Coverage = NewCoverage(undefined, unsupported, supported)

}

func (cs *CombinedSpec) GenerateReport(fileName string) error {
	t, err := template.ParseFiles("spec/combined_spec.html")
	if err != nil {
		return err
	}

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	cs.CalculateCoverage()

	err = t.ExecuteTemplate(w, "combined_spec", cs)
	if err != nil {
		return err
	}

	w.Flush()

	return nil

}

type Summary struct {
	OS            string
	Links         map[string]string
	SingleSpecs   []Spec
	CombinedSpecs []CombinedSpec
}

func (s *Summary) generateLinks(prefix string) {
	supportedOSs := [2]string{"darwin", "linux"}
	s.Links = make(map[string]string)

	for _, os := range supportedOSs {
		if os != s.OS {
			s.Links[os] = s.fileName(os)
		}
	}
}

func (s *Summary) fileName(variant string) string {
	return fmt.Sprintf("summary-%v.html", variant)
}

func NewSummary() *Summary {
	return &Summary{
		OS: runtime.GOOS,
	}
}

func (s *Summary) GenerateReport(dir string) error {
	s.generateLinks(dir)

	t, err := template.ParseFiles("spec/summary.html", "spec/combined_spec.html", "spec/spec.html")
	if err != nil {
		return err
	}

	fileName := filepath.Join(dir, s.fileName(runtime.GOOS))
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer f.Close()
	w := bufio.NewWriter(f)

	err = t.ExecuteTemplate(w, "summary", s)
	if err != nil {
		return err
	}

	w.Flush()

	return nil

}
