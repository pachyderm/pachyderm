package main

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	projectName = "ci-metrics"
	repoName    = "go-test-results-raw"
	fileSuffix  = "go-test-results.jsonl"
)

var pachdAddress = findPachdAddress()
var logger = log.Default()

func findPachdAddress() string {
	env := os.Getenv("OPS_PACHD_ADDRESS")
	if len(env) > 0 {
		return env
	}
	return "grpcs://pachyderm.pachops.com:443"
}

func sanitizeName(s string) string {
	ret := []byte(strings.ReplaceAll(s, " ", "_"))
	invalidCharacters := regexp.MustCompile("[^a-zA-Z0-9-_]+")
	ret = invalidCharacters.ReplaceAll(ret, []byte("-"))
	return string(ret)
}

func findDestinationPath(path string, branch string, jobName string, resultsFolder string) string {
	tmpPath := strings.Split(strings.TrimPrefix(path, resultsFolder), "/") // remove standard /tmp/test-results/ folder
	tmpPath = append([]string{branch, jobName}, tmpPath...)                // set up output file location
	return filepath.Join(tmpPath...)
}

func main() {
	logger.Println("Beginning results upload.")
	robotToken := os.Getenv("PACHOPS_PACHYDERM_ROBOT_TOKEN")
	if len(robotToken) == 0 {
		logger.Println("No pachyderm robot token found. Continuing with unauthenticated pach client.")
	}
	resultsFolder := os.Getenv("TEST_RESULTS")
	if len(resultsFolder) == 0 {
		logger.Fatal("TEST_RESULTS needs to be populated to find the test results folder.")
	}
	var branchFolderName string
	if branch := sanitizeName(os.Getenv("CIRCLE_BRANCH")); len(branch) != 0 {
		branchFolderName = branch
	} else if tag := sanitizeName(os.Getenv("CIRCLE_TAG")); len(tag) != 0 {
		branchFolderName = tag
	} else {
		logger.Fatal("CIRCLE_BRANCH or CIRCLE_TAG needs to be populated to upload test results.")
	}
	jobName := sanitizeName(os.Getenv("CIRCLE_JOB"))
	if len(jobName) == 0 {
		logger.Fatal("CIRCLE_JOB needs to be populated to upload test results.")
	}

	pachClient, err := client.NewFromURIContext(context.Background(), pachdAddress)
	if err != nil {
		logger.Fatalf("Provisioning pach client: %v", err)
	}
	pachClient.SetAuthToken(robotToken)

	eg, _ := errgroup.WithContext(pachClient.Ctx())
	err = filepath.WalkDir(resultsFolder, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(d.Name(), fileSuffix) {
			eg.Go(func() error {
				file, err := os.Open(path)
				if err != nil {
					return errors.Wrapf(err, "opening file %v", d.Name())
				}
				defer file.Close()

				destPath := findDestinationPath(path, branchFolderName, jobName, resultsFolder)
				commit := client.NewProjectCommit(projectName, repoName, "master", "")
				err = pachClient.PutFile(commit, destPath, file)
				if err != nil {
					return errors.Wrapf(err, "putting file: %v to commit %v", destPath, commit.GetID())
				}
				logger.Printf("Succeeded uploading %s \n", path)
				return nil
			})
		}
		return nil
	})
	if err != nil {
		logger.Fatalf("Failed walking file paths: %v", err)
	}
	if goRoutineErrs := eg.Wait(); goRoutineErrs != nil {
		logger.Fatalf("Failed creating files: %v", err)
	}
	logger.Println("Successfully uploaded files.")
}
