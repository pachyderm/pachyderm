package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
	"k8s.io/kubectl/pkg/util/slice"
)

type testOutput struct {
	Time    time.Time
	Action  string
	Package string
	Output  string
}

func main() {
	log.InitPachctlLogger()
	ctx := pctx.Background("test-collector")
	tags := flag.String("tags", "", "Tags to run, for example k8s. Tests without this flag will not be selected.")
	fileName := flag.String("file", "tests_to_run.csv", "Tags to run, for example k8s. Tests without this flag will not be selected.")
	threadPool := flag.Int("threads", 2, "Number of packages to collect tests from concurrently.")
	flag.Parse()
	err := run(ctx, *tags, *fileName, *threadPool)
	if err != nil {
		log.Error(ctx, "Error during tests splitting", zap.Error(err))
		os.Exit(1)
	}
	os.Exit(0)
}

func run(ctx context.Context, tags string, fileName string, threadPool int) error {
	tagsArg := ""
	var testIdsUntagged map[string][]string
	if tags != "" {
		tagsArg = fmt.Sprintf("-tags=%s", tags)
		var err error
		testIdsUntagged, err = testNames(ctx, "./...", "") // collect for set difference
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
	testIdsTagged, err := testNames(ctx, "./...", tagsArg)
	if err != nil {
		return errors.EnsureStack(err)
	}
	var testIds map[string][]string

	if tags != "" {
		// set difference to get ONLY tagged tests
		testIds = subtractTestSet(testIdsTagged, testIdsUntagged)
	} else {
		testIds = testIdsTagged
	}
	log.Info(ctx, "tests and packages collected", zap.Any("tests", testIds))
	err = outputToFile(fileName, testIds)
	return err
}

func subtractTestSet(testIdsTagged map[string][]string, testIdsUntagged map[string][]string) map[string][]string {
	testIds := map[string][]string{}
	for pkg, testsTagged := range testIdsTagged {
		testsUntagged, ok := testIdsUntagged[pkg] // DNJ TODO re-read and clean
		for _, testT := range testsTagged {
			if !ok || !slice.ContainsString(testsUntagged, testT, func(s string) string { return s }) {
				if _, ok := testIds[pkg]; !ok {
					testIds[pkg] = []string{testT}
				} else {
					testIds[pkg] = append(testIds[pkg], testT)
				}
			}
		}

	}
	return testIds
}

func testNames(ctx context.Context, pkg string, addtlCmdArgs ...string) (map[string][]string, error) {
	findTestArgs := append([]string{"test", pkg, "-json", "-list=."}, addtlCmdArgs...)
	cmd := exec.Command("go", findTestArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	scanner := bufio.NewScanner(cmdReader)
	err = cmd.Start()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	var testNames = map[string][]string{}
	for scanner.Scan() {
		testInfo := &testOutput{}
		raw := scanner.Bytes()
		if !bytes.HasPrefix(raw, []byte("{")) {
			continue // dependency download junk got shared
		}
		err := json.Unmarshal(raw, testInfo)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing json: %s", string(raw))
		}
		if testInfo.Action == "output" {
			output := strings.Trim(testInfo.Output, "\n ")
			if output != "" && !strings.HasPrefix(output, "Benchmark") &&
				!strings.HasPrefix(output, "ExampleAPIClient_") && // DNJ TODO should we be ignoring files or what here? ExampleChild() does currently run and would be missed
				!strings.HasPrefix(output, "? ") &&
				!strings.HasPrefix(output, "ok ") {
				if _, ok := testNames[testInfo.Package]; !ok {
					testNames[testInfo.Package] = []string{output}
				} else {
					testNames[testInfo.Package] = append(testNames[testInfo.Package], output)
				}
			}

		}

	}
	stderrBytes, err := io.ReadAll(stderr)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	errorText := string(stderrBytes)
	if errorText != "" {
		log.Error(ctx, string(errorText))
	}

	err = cmd.Wait()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return testNames, nil
}

func outputToFile(fileName string, pkgTests map[string][]string) error {
	// create a lock file so tests know to wait to start if running tests at same time
	lockFileName := fmt.Sprintf("lock-%s", fileName) // DNJ TODO share lock file name
	lockF, err := os.Create(lockFileName)
	if err != nil {
		return errors.EnsureStack(err)
	}
	err = lockF.Close()
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer os.Remove(lockFileName)

	f, err := os.Create(fileName)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for pkg, tests := range pkgTests {
		for _, test := range tests {
			_, err := w.WriteString(fmt.Sprintf("%s,%s\n", pkg, test))
			if err != nil {
				return errors.EnsureStack(err)
			}
		}
	}
	return errors.EnsureStack(w.Flush())
}
