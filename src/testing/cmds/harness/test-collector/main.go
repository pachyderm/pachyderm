package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const tagsExcludeAllFilesErr = "build constraints exclude all Go files"

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
	if tags != "" {
		tagsArg = fmt.Sprintf("-tags=%s", tags)
	}
	log.Info(ctx, "Collecting packages")
	pkgs, err := packageNames(tagsArg)
	if err != nil {
		return errors.EnsureStack(err)
	}
	testIds := map[string][]string{}
	testIdsMu := sync.Mutex{}
	sem := semaphore.NewWeighted(int64(threadPool))
	eg, _ := errgroup.WithContext(ctx)
	for _, pkg := range pkgs {
		loopPkg := pkg
		if loopPkg != "" && !strings.HasPrefix(loopPkg, "go: downloading") {
			err = sem.Acquire(ctx, 1)
			if err != nil {
				return errors.EnsureStack(err)
			}
			eg.Go(func() error { // DNJ TODO cleanup
				defer sem.Release(1)
				log.Info(ctx, "Collecting tests from package", zap.String("package", loopPkg))
				var testsToRun []string
				testsTagged, err := testNames(ctx, loopPkg, tagsArg)
				if err != nil {
					return errors.EnsureStack(err)
				}
				// DNJ TOD - cleanup/refactor
				if tagsArg == "" { // no tag means both test sets are the same, don't do subtraction
					testsToRun = testsTagged
				} else {
					testsToRun = []string{}
					testsUntagged, err := testNames(ctx, loopPkg)
					if err != nil {
						return errors.EnsureStack(err)
					}
					for _, testTagged := range testsTagged { // set subtraction to find exclusivly tagged tests
						found := false
						for _, testUntagged := range testsUntagged {
							if testTagged == testUntagged {
								found = true
								break
							}
						}
						if !found {
							testsToRun = append(testsToRun, testTagged)
						}
					}
				}
				if len(testsToRun) > 0 {
					testIdsMu.Lock()
					testIds[loopPkg] = testsToRun
					testIdsMu.Unlock()
				}
				return nil
			})
		}
	}
	if err := eg.Wait(); err != nil {
		return errors.EnsureStack(err)
	}
	outputToFile(fileName, testIds)
	return nil
}
func packageNames(addtlCmdArgs ...string) ([]string, error) {
	findPkgArgs := append([]string{"list"}, addtlCmdArgs...)
	findPkgArgs = append(findPkgArgs, "./...") // DNJ TODO - add packages flag
	pkgsOutput, err := exec.Command("go", findPkgArgs...).CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, string(pkgsOutput))
	}
	return strings.Split(string(pkgsOutput), "\n"), nil
}

func testNames(ctx context.Context, pkg string, addtlCmdArgs ...string) ([]string, error) {
	findTestArgs := append([]string{"test", pkg, "-list=."}, addtlCmdArgs...)
	cmd := exec.Command("go", findTestArgs...)
	log.Info(ctx, "About to run command to find tests ", zap.String("command", cmd.String()))
	testsOutputBytes, err := cmd.CombinedOutput()
	testsOutput := string(testsOutputBytes)
	if err != nil && !strings.Contains(testsOutput, tagsExcludeAllFilesErr) {
		return nil, errors.Wrapf(err, "Output from test list command: %s", testsOutput)
	}
	// Note that this includes k8s and non-k8s tests since tags are inclusive
	testList := strings.Split(testsOutput, "\n")
	var testNames = []string{}
	for _, test := range testList {
		if test != "" && !strings.HasPrefix(test, "Benchmark") &&
			!strings.HasPrefix(test, "ExampleAPIClient_") && // DNJ TODO should we be ignoring files or what here? ExampleChild() does currently run and would be missed
			!strings.HasPrefix(test, "? ") &&
			!strings.HasPrefix(test, "ok ") &&
			!strings.HasPrefix(test, "go: downloading") {
			testNames = append(testNames, strings.TrimSpace(test))
		}
	}
	return testNames, nil
}

func outputToFile(fileName string, pkgTests map[string][]string) error {
	// create a lock file so tests know to wait to start if running tests at same time
	lockFileName := fmt.Sprintf("lock-%s", fileName) // DNJ TODO share lock file name
	lockF, err := os.Create(lockFileName)
	if err != nil {
		return err
	}
	err = lockF.Close()
	if err != nil {
		return err
	}
	defer os.Remove(lockFileName)

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for pkg, tests := range pkgTests {
		for _, test := range tests {
			_, err := w.WriteString(fmt.Sprintf("%s,%s\n", pkg, test))
			if err != nil {
				return err
			}
		}
	}
	err = w.Flush()
	return err
}
