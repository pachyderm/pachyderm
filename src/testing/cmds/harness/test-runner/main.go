package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/shlex"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

func main() {
	log.InitPachctlLogger()
	ctx := pctx.Background("test-runner")
	tags := flag.String("tags", "", "Tags to run, for example k8s. Tests without this flag will not be selected.")
	fileName := flag.String("file", "tests_to_run.csv", "Tags to run, for example k8s. Tests without this flag will not be selected.")
	gotestsumArgsRaw := flag.String("gotestsum-args", "", "Additional arguments to pass to the gotestsum portion of the test command.")
	gotestArgsRaw := flag.String("gotest-args", "", "Additional arguments to pass to the 'go test' portion of the test command.")
	shard := flag.Int("shard", 0, "0 indexed current runner index that we are on.")
	totalShards := flag.Int("total-shards", 1, "Total number of runners that we are sharding over.")
	threadPool := flag.Int("threads", 1, "Number of tests to execute concurrently.")
	flag.Parse()
	gotestsumArgs, err := shlex.Split(*gotestsumArgsRaw)
	if err != nil {
		log.Error(ctx, "Error parsing gotestsumArgs", zap.Error(err))
		os.Exit(1)
	}
	gotestArgs, err := shlex.Split(*gotestArgsRaw)
	if err != nil {
		log.Error(ctx, "Error parsing gotestArgs", zap.Error(err))
		os.Exit(1)
	}

	err = run(ctx,
		*tags,
		*fileName,
		gotestsumArgs,
		gotestArgs,
		*shard,
		*totalShards,
		*threadPool,
	)
	if err != nil {
		log.Error(ctx, "Error running tests", zap.Error(err))
		os.Exit(1)
	}
	os.Exit(0)
}

func run(ctx context.Context, tags string, fileName string, gotestsumArgs []string, gotestArgs []string, shard int, totalShards int, threadPool int) error {
	tests, err := readTests(ctx, fileName)
	if err != nil {
		return errors.Wrapf(err, "reading file %v", fileName)
	}
	slices.Sort(tests) // sort so we shard them from a pre-determined order

	// loop through by the number of shards so that each gets a roughly equal number on each
	testsForShard := map[string][]string{}
	for idx := shard; idx < len(tests); idx += totalShards {
		val := strings.Split(tests[idx], ",")
		pkg := val[0]
		testName := val[1]
		// index all tests by package as we collect the ones for this shard. This lets
		// us run all tests in each package on this shard with one `go test` command, preserving the serial
		// running of tests without t.parallel the same way that go test ./... would since
		// got test also runs packages in paralllel.
		if _, ok := testsForShard[pkg]; !ok {
			testsForShard[pkg] = []string{testName}
		} else {
			testsForShard[pkg] = append(testsForShard[pkg], testName)
		}
	}

	// DNJ TODO - incorporate gomaxprocs or add parallel parameter

	eg, _ := errgroup.WithContext(ctx)
	eg.SetLimit(threadPool)
	for pkg, tests := range testsForShard {
		threadLocalPkg := pkg
		threadLocalTests := tests
		if err != nil {
			return errors.EnsureStack(err)
		}
		eg.Go(func() error {
			return errors.EnsureStack(runTest(threadLocalPkg, threadLocalTests, tags, gotestsumArgs, gotestArgs))
		})
	}
	err = eg.Wait()

	if err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

func readTests(ctx context.Context, fileName string) ([]string, error) {
	lockFileName := fmt.Sprintf("lock-%s", fileName)
	err := backoff.RetryNotify(func() error {
		if _, err := os.Stat(fileName); err != nil {
			return errors.EnsureStack(err) // couldn't read file, so retry until it can
		}
		if _, err := os.Stat(lockFileName); err == nil {
			return errors.Errorf("lock file for test collection still exists")
		}
		return nil
	}, backoff.NewConstantBackOff(time.Second*5).For(time.Minute*20), func(err error, d time.Duration) error {
		log.Info(ctx, "retry waiting for tests to be collected.", zap.Error(err))
		return nil
	})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	tests := []string{}
	file, err := os.Open(fileName)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tests = append(tests, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return tests, nil
}

// run tests with `go test`. We run one package at a time so tests with the same name in different packages
// like TestConfig or TestDebug don't run multiple times if they land on separate shards.
func runTest(pkg string, testNames []string, tags string, gotestsumArgs []string, gotestArgs []string) error {
	resultsFolder := os.Getenv("TEST_RESULTS")
	pkgShort := strings.ReplaceAll(strings.TrimPrefix(pkg, "github.com/pachyderm/pachyderm/v2"), "/", "-")
	runTestArgs := []string{
		fmt.Sprintf("--packages=%s", pkg),
		"--rerun-fails",
		"--rerun-fails-max-failures=1",
		"--format=testname",
		"--debug",
		fmt.Sprintf("--junitfile=%s/circle/gotestsum-report-%s.xml", resultsFolder, pkgShort),
		fmt.Sprintf("--jsonfile=%s/go-test-results-%s.jsonl", resultsFolder, pkgShort),
	}
	if len(gotestsumArgs) > 0 {
		runTestArgs = append(runTestArgs, gotestsumArgs...)
	}
	testRegex := strings.Builder{}
	for _, test := range testNames {
		if testRegex.Len() > 0 {
			testRegex.WriteString("|")
		}
		testRegex.WriteString(fmt.Sprintf("^%s$", test))
	}
	runTestArgs = append(runTestArgs, "--",
		fmt.Sprintf("-tags=%s", tags),
		fmt.Sprintf("-run=%s", testRegex.String()),
	)
	if len(gotestArgs) > 0 {
		runTestArgs = append(runTestArgs, gotestArgs...)
	}

	cmd := exec.Command("gotestsum", runTestArgs...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "CGO_ENABLED=0", "GOCOVERDIR=\"/tmp/test-results/\"") // DNJ TODO - parameter - how to take form args?
	fmt.Printf("Running command %v\n", cmd.String())
	testsOutput, err := cmd.CombinedOutput()
	_, copyErr := io.Copy(os.Stdout, strings.NewReader(string(testsOutput)))
	if err != nil {
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(copyErr)

}
