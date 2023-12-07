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
	log.InitBatchLogger("test-collector.log")
	ctx := pctx.Background("test-collector")
	tags := flag.String("tags", "", "Tags to run, for example k8s. Tests without this flag will not be selected.")
	exclusiveTags := flag.Bool("exclusiveTags", true, "If true, ONLY tests with the specified tags will run. If false, "+
		"the default behavior of 'go test' of including tagged and untagged tests is used.")
	fileName := flag.String("file", "tests_to_run.csv", "Output file listing the packages and tests to run. Used by the runner script.")
	pkg := flag.String("pkg", "./...", "Package to run defaults to all packages.")
	threadPool := flag.Int("threads", 2, "Number of packages to collect tests from concurrently.")
	flag.Parse()
	err := run(ctx, *tags, *exclusiveTags, *fileName, *pkg, *threadPool)
	if err != nil {
		log.Exit(ctx, "Error during tests splitting", zap.Error(err))
	}
	os.Exit(0)
}

func run(ctx context.Context, tags string, exclusiveTags bool, fileName string, pkg string, threadPool int) error {
	var tagsArg string
	if tags != "" {
		tagsArg = fmt.Sprintf("-tags=%s", tags)
	}
	testIdsTagged, err := testNames(ctx, pkg, tagsArg)
	if err != nil {
		return errors.EnsureStack(err)
	}
	var testIds map[string][]string
	if exclusiveTags && tags != "" {
		// set difference to get ONLY tagged tests
		var err error
		testIdsUntagged, err := testNames(ctx, pkg, "") // collect for set difference
		if err != nil {
			return errors.EnsureStack(err)
		}
		testIds = subtractTestSet(testIdsTagged, testIdsUntagged)
	} else {
		testIds = testIdsTagged
	}
	log.Info(ctx, "tests and packages collected", zap.Any("tests", testIds))
	err = outputToFile(fileName, testIds)
	return err
}

// get tests that are tagged, but not in the untagged list since that is inclusive
func subtractTestSet(testIdsTagged map[string][]string, testIdsUntagged map[string][]string) map[string][]string {
	resultTests := map[string][]string{}
	for pkg, testsTagged := range testIdsTagged {
		testsUntagged, ok := testIdsUntagged[pkg]
		for _, testNameTagged := range testsTagged {
			if !ok || !slice.ContainsString(testsUntagged, testNameTagged, func(s string) string { return s }) {
				if _, ok := resultTests[pkg]; !ok {
					resultTests[pkg] = []string{testNameTagged}
				} else {
					resultTests[pkg] = append(resultTests[pkg], testNameTagged)
				}
			}
		}

	}
	return resultTests
}

func testNames(ctx context.Context, pkg string, addtlCmdArgs ...string) (map[string][]string, error) {
	findTestArgs := append([]string{"test", pkg, "-json", "-list=."}, addtlCmdArgs...)
	cmd := exec.Command("go", findTestArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	cmd.Stderr = log.WriterAt(log.ChildLogger(ctx, "stderr"), log.InfoLevel)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOMAXPROCS=16") // This prevents the command from running wild eating up processes in the pipelines
	err = cmd.Start()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	testNames, err := readTests(stdout)
	if err != nil {
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return testNames, nil
}

func readTests(stdout io.Reader) (map[string][]string, error) {
	var testNames = map[string][]string{}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		testInfo := &testOutput{}
		raw := scanner.Bytes()
		if !bytes.HasPrefix(raw, []byte("{")) { // dependency download junk got shared
			continue
		}
		if err := json.Unmarshal(raw, testInfo); err != nil {
			return nil, errors.Wrapf(err, "parsing json: %s", string(raw))
		}
		if testInfo.Action == "output" {
			output := strings.Trim(testInfo.Output, "\n ")
			if output != "" && !strings.HasPrefix(output, "Benchmark") &&
				!strings.HasPrefix(output, "ExampleAPIClient_") &&
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
	return testNames, nil
}

func outputToFile(fileName string, pkgTests map[string][]string) error {
	tmpFileName := fmt.Sprintf("%s.tmp", fileName)
	f, err := os.Create(tmpFileName)
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
	err = errors.EnsureStack(w.Flush())
	if err != nil {
		return err
	}
	err = os.Rename(tmpFileName, fileName)
	return errors.EnsureStack(err)
}
