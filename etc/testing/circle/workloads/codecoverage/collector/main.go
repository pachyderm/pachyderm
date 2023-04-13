package main

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	projectName = "ci-metrics"
	repoName    = "codecoverage-source"
)

var pachdAddress = findPachdAddress()
var logger = log.Default()

func findPachdAddress() string {
	env := os.Getenv("COV_PACHD_ADDRESS")
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

func main() {
	ctx := context.Background()
	robotToken := os.Getenv("PACHOPS_PACHYDERM_ROBOT_TOKEN")
	coverageFolder := os.Getenv("TEST_RESULTS")
	if len(coverageFolder) == 0 {
		logger.Fatal("TEST_RESULTS needs to be populated to find the test results folder.")
	}
	branch := sanitizeName(os.Getenv("CIRCLE_BRANCH"))
	if len(branch) == 0 {
		logger.Fatal("CIRCLE_BRANCH needs to be populated to upload code coverage.")
	}
	jobName := sanitizeName(os.Getenv("CIRCLE_JOB"))
	if len(jobName) == 0 {
		logger.Fatal("CIRCLE_JOB needs to be populated to upload code coverage.")
	}

	pachClient, err := client.NewFromURI(pachdAddress)
	if err != nil {
		logger.Fatal(err)
	}
	pachClient.SetAuthToken(robotToken)
	// repo := client.NewProjectRepo(projectName, repoName)
	// commit := repo.NewCommit("master", "") // DNJ TODO - how to handle branching?
	commit, err := pachClient.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
		Branch: client.NewProjectBranch(projectName, repoName, "master"),
	})
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("Successfully started commit.")
	var wg sync.WaitGroup // DNJ TODO - use error group now that we have a context?
	filepath.WalkDir(coverageFolder, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() { // DNJ TODO combine ifs?
			if strings.HasSuffix(d.Name(), "-cover.tgz") || strings.HasSuffix(d.Name(), "coverage.txt") {
				wg.Add(1)
				go func() {
					defer wg.Done()
					file, err := os.Open(path)
					if err != nil {
						logger.Fatal(err) // DNJ TODO - decide error handling after refactor
					}
					logger.Printf("Uploading file: %s ...", path)
					tmpPath := strings.Split(path, "/")
					tmpPath = tmpPath[3:]
					tmpPath = append([]string{branch, jobName}, tmpPath...) // DNJ TODO - should we delete folder on merge to master?
					defer file.Close()

					err = pachClient.PutFile(commit, filepath.Join(tmpPath...), file)
					if err != nil {
						logger.Fatal(err)
					}
					logger.Printf("Succeeded %s \n", path)
				}()
			}
		}
		return nil
	})
	wg.Wait()
	_, err = pachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{Commit: commit})
	if err != nil {
		logger.Fatal(err)
	} else {
		logger.Println("Successfully committed.")
	}

}
