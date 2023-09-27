package main

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

func main() {
	log.InitPachctlLogger()
	ctx := pctx.Background("")
	err := run(ctx)
	if err != nil {
		log.Error(ctx, "Error during tests splitting", zap.Error(err))
	}
}

func run(ctx context.Context) error {
	testNamesK8s, err := testNames("-tags=k8s")
	if err != nil {
		return errors.EnsureStack(err)
	}
	testNamesUnit, err := testNames()
	if err != nil {
		return errors.EnsureStack(err)
	}
	testNamesOnlyK8s := []string{}
	for _, testK8s := range testNamesK8s { // set subtraction
		found := false
		for _, testUnit := range testNamesUnit {
			if testK8s == testUnit {
				found = true
				break // do not add since this test had no tag
			}
		}
		if !found {
			testNamesOnlyK8s = append(testNamesOnlyK8s, testK8s)
		}
	}
	outputToFile("tests_unit.txt", testNamesUnit)
	outputToFile("tests_k8s.txt", testNamesOnlyK8s)
	return nil
}

func testNames(addtlCmdArgs ...string) ([]string, error) {
	findTestArgs := append([]string{"test", "./src/...", "-list=."}, addtlCmdArgs...)
	testsOutput, err := exec.Command("go", findTestArgs...).CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, string(testsOutput))
	}
	// Note that this includes k8s and non-k8s tests since tags are inclusive
	testsK8s := strings.Split(string(testsOutput), "\n")
	var testNames = []string{}
	for _, test := range testsK8s {
		if !strings.HasPrefix(test, "Benchmark") &&
			!strings.HasPrefix(test, "?") &&
			!strings.HasPrefix(test, "ok") {
			testNames = append(testNames, test)
		}
	}
	return testNames, nil
}

func outputToFile(fileName string, testNames []string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, test := range testNames {
		_, err := w.WriteString(test + "\n")
		if err != nil {
			return err
		}
	}
	err = w.Flush()
	return err
}
