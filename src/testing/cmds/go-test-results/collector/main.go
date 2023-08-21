package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	gotestresults "github.com/pachyderm/pachyderm/v2/src/testing/cmds/go-test-results"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	projectName        = "ci-metrics"
	repoName           = "go-test-results-raw"
	fileSuffix         = "go-test-results.jsonl"
	commitRetryBackoff = 10 * time.Second
	commitMaxRetries   = 30
)

var pachdAddress = findPachdAddress()
var invalidCharacters = regexp.MustCompile("[^a-zA-Z0-9-_]+")
var repo = &pfs.Repo{
	Name: repoName,
	Type: pfs.UserRepoType,
	Project: &pfs.Project{
		Name: projectName,
	},
}
var commit = &pfs.Commit{
	Branch: &pfs.Branch{
		Name: "master",
		Repo: repo,
	},
}

func findPachdAddress() string {
	env := os.Getenv("OPS_PACHD_ADDRESS")
	if len(env) > 0 {
		return env
	}
	return "grpcs://pachyderm.pachops.com:443"
}

// This runs in circle-ci pipelines and sends the raw test data to pachyderm for transform.
func main() {
	log.InitPachctlLogger()
	ctx := pctx.Background("")
	err := run(ctx)
	if err != nil {
		log.Error(ctx, "Error during metric collection", zap.Error(err))
	}
}

func run(ctx context.Context) error {
	robotToken := os.Getenv("PACHOPS_PACHYDERM_ROBOT_TOKEN")
	if len(robotToken) == 0 {
		log.Info(ctx, "No pachyderm robot token found. Continuing with unauthenticated pach client.")
	}
	resultsFolder := os.Getenv("TEST_RESULTS")
	if len(resultsFolder) == 0 {
		return errors.WithStack(fmt.Errorf("TEST_RESULTS needs to be populated to find the test results folder."))
	}
	if _, err := os.Stat(resultsFolder); os.IsNotExist(err) {
		return errors.WithStack(fmt.Errorf("The test result folder at %v does not exist. Exiting early.", resultsFolder))
	}
	// connect and authenticate to pachyderm
	pachClient, err := client.NewFromURIContext(context.Background(), pachdAddress)
	if err != nil {
		return errors.Wrapf(err, "Problem provisioning pach client")
	}
	pachClient.SetAuthToken(robotToken)

	// upload general job information
	jobInfo, err := findJobInfoFromCI()
	if err != nil {
		return errors.Wrapf(err, "Problem collecting job info")
	}
	basePath := findBasePath(jobInfo)
	// upload individual test  results
	eg, _ := errgroup.WithContext(ctx)
	err = filepath.WalkDir(resultsFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), fileSuffix) {
			eg.Go(func() error {
				return uploadTestResult(path, d, basePath, resultsFolder, pachClient, commit)
			})
		}
		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "Problem walking file paths")
	}
	if goRoutineErrs := eg.Wait(); goRoutineErrs != nil {
		return errors.Wrapf(goRoutineErrs, "Problem putting files to pachyderm")
	}
	// do this last so the egress knows that we are done for this job/executor
	if err = uploadJobInfo(basePath, jobInfo, pachClient, commit); err != nil {
		return errors.Wrapf(err, "Problem uploading job info")
	}
	log.Info(ctx, "Successfully uploaded files.")
	return nil
}

func sanitizeName(s string) string {
	ret := bytes.ReplaceAll([]byte(s), []byte(" "), []byte("_"))
	ret = invalidCharacters.ReplaceAll(ret, []byte("-"))
	return string(ret)
}

func findDestinationPath(path string, statsPath string, resultsFolder string) string {
	tmpPath := strings.Split(strings.TrimPrefix(path, resultsFolder), "/") // remove standard /tmp/test-results/ folder
	tmpPath = append([]string{statsPath}, tmpPath...)                      // set up output file location
	return filepath.Join(tmpPath...)
}

func findJobInfoFromCI() (gotestresults.JobInfo, error) {
	jobInfo := gotestresults.JobInfo{}
	mapping := make(map[string](*string))
	mapping["CIRCLE_BRANCH"] = &jobInfo.Branch
	mapping["CIRCLE_TAG"] = &jobInfo.Tag
	mapping["CIRCLE_WORKFLOW_ID"] = &jobInfo.WorkflowId
	mapping["CIRCLE_WORKFLOW_JOB_ID"] = &jobInfo.JobId
	mapping["CIRCLE_JOB"] = &jobInfo.JobName
	mapping["CIRCLE_SHA1"] = &jobInfo.Commit
	mapping["CIRCLE_BRANCH"] = &jobInfo.Branch
	mapping["CIRCLE_USERNAME"] = &jobInfo.Username
	mapping["CIRCLE_PULL_REQUESTS"] = &jobInfo.PullRequests

	allowEmpty := make(map[string]bool)
	allowEmpty["CIRCLE_BRANCH"] = true
	allowEmpty["CIRCLE_TAG"] = true
	allowEmpty["CIRCLE_PULL_REQUESTS"] = true
	allowEmpty["CIRCLE_USERNAME"] = true

	for envVar, statsField := range mapping {
		*statsField = os.Getenv(envVar)
		if *statsField == "" && !allowEmpty[envVar] {
			return jobInfo, errors.EnsureStack(fmt.Errorf("%s needs to be populated to upload test results.", envVar))
		}
	}
	// non-string fields
	jobInfo.JobTimestamp = time.Now().UTC()
	var err error
	jobInfo.JobNumExecutors, err = strconv.Atoi(os.Getenv("CIRCLE_NODE_TOTAL"))
	if err != nil {
		return jobInfo, errors.Wrap(err, "Parsing CIRCLE_NODE_TOTAL")
	}
	jobInfo.JobExecutor, err = strconv.Atoi(os.Getenv("CIRCLE_NODE_INDEX"))
	return jobInfo, errors.Wrap(err, "Parsing CIRCLE_NODE_INDEX")
}

func findBasePath(jobInfo gotestresults.JobInfo) string {
	var branchFolderName string
	if jobInfo.Branch != "" {
		branchFolderName = jobInfo.Branch
	} else {
		branchFolderName = jobInfo.Tag
	}
	basePath := filepath.Join(sanitizeName(branchFolderName), sanitizeName(jobInfo.JobName), strconv.Itoa(jobInfo.JobExecutor))
	return basePath
}

// Uploads job info to pachyderm cluster.
func uploadJobInfo(basePath string, jobInfo gotestresults.JobInfo, pachClient *client.APIClient, commit *pfs.Commit) error {
	jobInfoJson, err := json.Marshal(jobInfo)
	if err != nil {
		return errors.Wrapf(err, "Could not marshal CI job stats to json: %v ", jobInfo)
	}
	if err = pachClient.PutFile(commit, filepath.Join(basePath, "JobInfo.json"), bytes.NewReader(jobInfoJson)); err != nil {
		return errors.Wrapf(err, "Could not output CI job stats to pachyderm cluster ")
	}
	return nil
}

func uploadTestResult(path string, d fs.DirEntry, basePath string, resultsFolder string, pachClient *client.APIClient, commit *pfs.Commit) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "opening file %v", d.Name())
	}
	defer file.Close()
	destPath := findDestinationPath(path, basePath, resultsFolder)
	if err = pachClient.PutFile(commit, destPath, file); err != nil {
		return errors.Wrapf(err, "putting file: %v to commit %v", destPath, commit.GetId())
	}
	return nil
}
