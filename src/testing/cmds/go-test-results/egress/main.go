package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	gotestresults "github.com/pachyderm/pachyderm/v2/src/testing/cmds/go-test-results"
	"github.com/pkg/errors"
)

var logger = log.Default()
var postgresqlUser = os.Getenv("POSTGRESQL_USER")
var postgresqlPassword = os.Getenv("POSTGRESQL_PASSWORD")
var postgresqlHost = os.Getenv("POSTGRESQL_HOST")

const (
	jobInfoFileName string = "JobInfo.json"
)

type TestResultLine struct {
	Time     time.Time `json:"Time"`
	Action   string    `json:"Action"`
	TestName string    `json:"Test"`
	Package  string    `json:"Package"`
	Elapsed  float32   `json:"Elapsed"`
	Output   string    `json:"Output"`
}

// This is built and runs in a pachyderm pipeline to egress data uploaded to pachyderm to postgresql.
func main() {
	logger.Println("Beginning egress transform of data to sql DB")
	inputFolder := os.Args[1]

	db, err := sqlx.Open("pgx", fmt.Sprintf("user=%s password=%s host=%s port=5432 database=ci_metrics", postgresqlUser, postgresqlPassword, postgresqlHost))
	if err != nil {
		logger.Fatalf("Failed opening db connection %v", err)
	}
	// filpathWalkDir does not evaluate the PFS symlink
	sym, err := filepath.EvalSymlinks(inputFolder)
	if err != nil {
		logger.Fatal(err)
	}
	jobInfoPaths := make(map[string]gotestresults.JobInfo)
	var testResultPaths []string
	err = filepath.WalkDir(sym, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == jobInfoFileName {
			jobInfo, err := readJobInfo(path)
			if err != nil {
				return err
			}
			return insertJobInfo(jobInfo, jobInfoPaths, path, d, db)
		} else {
			// we need the job info inserted first due to foreign key constraints,
			// so save results to loop over later ensuring ordered inserts
			testResultPaths = append(testResultPaths, path)
		}
		return nil
	})
	if err != nil {
		logger.Fatal("Problem walking file tree for job info ", err)
	}
	if len(jobInfoPaths) == 0 {
		logger.Println("No new job info found, exiting without inserting test results.")
		return
	}
	for _, path := range testResultPaths {
		if err = insertTestResultFile(path, jobInfoPaths, db); err != nil {
			logger.Fatal("Problem inserting test results into DB ", err)
		}
	}
}

func readJobInfo(path string) (*gotestresults.JobInfo, error) {
	jobInfoFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshalling job info json ")
	}
	defer jobInfoFile.Close()
	jobInfoJson, err := io.ReadAll(jobInfoFile)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshalling job info json ")
	}
	var jobInfo gotestresults.JobInfo
	if err = json.Unmarshal(jobInfoJson, &jobInfo); err != nil {
		return nil, errors.Wrapf(err, "unmarshalling job info json ")
	}
	return &jobInfo, nil
}

func insertJobInfo(jobInfo *gotestresults.JobInfo, jobInfoPaths map[string]gotestresults.JobInfo, path string, d fs.DirEntry, db *sqlx.DB) error {
	_, err := db.Exec(`INSERT INTO public.ci_jobs (
							workflow_id, 
							job_id, 
							job_executor,
							job_name, 
							job_timestamp, 
							job_num_executors,
							commit_sha, 
							branch, 
							tag,
							pull_requests
						) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		jobInfo.WorkflowId,
		jobInfo.JobId,
		jobInfo.JobExecutor,
		jobInfo.JobName,
		jobInfo.JobTimestamp,
		jobInfo.JobNumExecutors,
		jobInfo.Commit,
		jobInfo.Branch,
		jobInfo.Tag,
		jobInfo.PullRequests,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Unique key constraint violation. In other words, the record already exists.
			// Ignore insert to since we don't want to change the timestamp, but log it.
			logger.Printf("Job info already inserted, Skipping insert for workflow %v  job %v  executor %v\n", jobInfo.WorkflowId, jobInfo.JobId, jobInfo.JobExecutor)
			return nil
		}
		return errors.Wrapf(err, "inserting job info for workflow %v  job %v  executor %v", jobInfo.WorkflowId, jobInfo.JobId, jobInfo.JobExecutor)
	}
	jobInfoPaths[strings.TrimSuffix(path, d.Name())] = *jobInfo // read for use when inserting individual test results, don't insert if errored
	logger.Printf("Inserted job info for workflow %v job %v executor %v", jobInfo.WorkflowId, jobInfo.JobId, jobInfo.JobExecutor)
	return nil
}

func insertTestResultFile(path string, jobInfoPaths map[string]gotestresults.JobInfo, db *sqlx.DB) error {
	resultsFile, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "opening results file ")
	}
	defer resultsFile.Close()

	splitPath := strings.Split(path, "/")
	fileName := splitPath[len(splitPath)-1]
	resultsScanner := bufio.NewScanner(resultsFile)
	for resultsScanner.Scan() {
		resultJson := resultsScanner.Bytes()
		var result TestResultLine
		err = json.Unmarshal(resultJson, &result)
		if err != nil {
			logger.Printf("Failed unmarshalling file as a test result - %s - %v. Continuing to parse results file.\n", path, err)
			continue
		}
		if result.Action == "output" {
			continue // output is mostly logs and coverage and it's not that useful, so we don't store output-only actions
		}
		// get job info previously inserted in order to link foreign keys to jobs table
		jobInfo, ok := jobInfoPaths[strings.TrimSuffix(path, fileName)]
		if !ok {
			return errors.WithStack(fmt.Errorf("Failed to find job info for %v - file name %v - job infos %v ", path, fileName, jobInfoPaths))
		}
		_, err = db.Exec(`INSERT INTO public.test_results (
									id, 
									workflow_id, 
									job_id, 
									job_executor,
									test_name, 
									package, 
									"action", 
									elapsed_seconds, 
									"output"
								) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(),
			jobInfo.WorkflowId,
			jobInfo.JobId,
			jobInfo.JobExecutor,
			result.TestName,
			result.Package,
			result.Action,
			result.Elapsed,
			result.Output,
		)
		if err != nil {
			logger.Printf("Failed updating SQL DB %s - %v - continuing to parse results file.\n", string(resultJson), err)
		}
	}
	return nil
}
