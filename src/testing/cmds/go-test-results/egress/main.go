package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"

	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	gotestresults "github.com/pachyderm/pachyderm/v2/src/testing/cmds/go-test-results"
)

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
	log.InitPachctlLogger()
	var ctx = pctx.Background("")
	log.Info(ctx, "Running DB Migrate")
	out, err := exec.Command("tern", "migrate").CombinedOutput()
	if err != nil {
		log.Exit(ctx, "Error running migrates.", zap.ByteString("Migrate output", out), zap.Error(err))
	}
	log.Info(ctx, "Migrate successful, beginning egress transform of data to sql DB")
	inputFolder := os.Args[1]
	db, err := dbutil.NewDB(
		dbutil.WithHostPort(postgresqlHost, 5432),
		dbutil.WithDBName("ci_metrics"),
		dbutil.WithUserPassword(postgresqlUser, postgresqlPassword),
	)
	// db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		log.Exit(ctx, "Failed opening db connection %v", zap.Error(err))
	}
	// filpathWalkDir does not evaluate the PFS symlink
	sym, err := filepath.EvalSymlinks(inputFolder)
	if err != nil {
		log.Exit(ctx, "Problem evaluating symlinks", zap.Error(err))
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
			return insertJobInfo(ctx, jobInfo, jobInfoPaths, path, d, db)
		} else {
			// we need the job info inserted first due to foreign key constraints,
			// so save results to loop over later ensuring ordered inserts
			testResultPaths = append(testResultPaths, path)
		}
		return nil
	})
	if err != nil {
		log.Exit(ctx, "Problem walking file tree for job info ", zap.Error(err))
	}
	if len(jobInfoPaths) == 0 {
		log.Info(ctx, "No new job info found, exiting without inserting test results.")
		return
	}
	for _, path := range testResultPaths {
		if err = insertTestResultFile(ctx, path, jobInfoPaths, db); err != nil {
			log.Exit(ctx, "Problem inserting test results into DB ", zap.Error(err))
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

func insertJobInfo(
	ctx context.Context,
	jobInfo *gotestresults.JobInfo,
	jobInfoPaths map[string]gotestresults.JobInfo,
	path string,
	d fs.DirEntry,
	db *sqlx.DB,
) error {
	_, err := db.NamedExec(`INSERT INTO public.ci_jobs (
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
						) VALUES (:workflowid, :jobid, :jobexecutor, :jobname, :jobtimestamp, :jobnumexecutors, :commit, :branch, :tag, :pullrequests)`,
		jobInfo,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Unique key constraint violation. In other words, the record already exists.
			// Ignore insert to since we don't want to change the timestamp, but log it.
			log.Info(ctx, "Job info already inserted, Skipping insert for workflow %v  job %v  executor %v\n", zap.String("Workflow id", jobInfo.WorkflowId), zap.String("Job id", jobInfo.JobId), zap.Int("Job Executor", jobInfo.JobExecutor))
			return nil
		}
		return errors.Wrapf(err, "inserting job info for workflow %v  job %v  executor %v", jobInfo.WorkflowId, jobInfo.JobId, jobInfo.JobExecutor)
	}
	jobInfoPaths[strings.TrimSuffix(path, d.Name())] = *jobInfo // read for use when inserting individual test results, don't insert if errored
	log.Info(ctx, "Inserted job info for workflow %v job %v executor %v", zap.String("Workflow id", jobInfo.WorkflowId), zap.String("Job id", jobInfo.JobId), zap.Int("Job Executor", jobInfo.JobExecutor))
	return nil
}

func insertTestResultFile(ctx context.Context, path string, jobInfoPaths map[string]gotestresults.JobInfo, db *sqlx.DB) error {
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
			log.Info(ctx, "Failed unmarshalling file as a test result - %s - %v. Continuing to parse results file.\n", zap.String("path", path), zap.Error(err))
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
			log.Info(ctx, "Failed updating SQL DB %s - %v - continuing to parse results file.\n", zap.ByteString("result JSON", (resultJson)), zap.Error(err))
		}
	}
	return nil
}
