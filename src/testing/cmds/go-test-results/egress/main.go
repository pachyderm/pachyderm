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
	"golang.org/x/sync/errgroup"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	gotestresults "github.com/pachyderm/pachyderm/v2/src/testing/cmds/go-test-results"
)

const (
	jobInfoFileName      string = "JobInfo.json"
	uploadGoroutineLimit int    = 10
	uploadChunkSize      int    = 50
	testResultFieldCount int    = 9
)

var startTime time.Time

type TestResultLine struct {
	Time     time.Time `json:"Time"`
	Action   string    `json:"Action"`
	TestName string    `json:"Test"`
	Package  string    `json:"Package"`
	Elapsed  float32   `json:"Elapsed"`
	Output   string    `json:"Output"`
}

type ResultRow struct {
	Id             string  `db:"id"`
	WorkflowId     string  `db:"workflowid"`
	JobId          string  `db:"jobid"`
	JobExecutor    int     `db:"jobexecutor"`
	TestName       string  `db:"testname"`
	Package        string  `db:"package"`
	Action         string  `db:"action"`
	ElapsedSeconds float32 `db:"elapsedseconds"`
	Output         string  `db:"output"`
}

// This is built and runs in a pachyderm pipeline to egress data uploaded to pachyderm to postgresql.
func main() {
	startTime = time.Now()
	log.InitPachctlLogger()
	ctx := pctx.Background("")
	err := run(ctx)
	if err != nil {
		log.Exit(ctx, "Error during metric collection", zap.Error(err))
	}
}
func run(ctx context.Context) error {
	log.Info(ctx, "Running DB Migrate")
	out, err := exec.Command("tern", "migrate").CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "running migrates. %v", string(out))
	}
	log.Info(ctx, "Migrate successful, beginning egress transform of data to sql DB")
	inputFolder := os.Args[1]
	db, err := dbutil.NewDB(
		dbutil.WithHostPort(gotestresults.PostgresqlHost, 5432),
		dbutil.WithDBName("ci_metrics"),
		dbutil.WithUserPassword(gotestresults.PostgresqlUser, gotestresults.PostgresqlPassword),
	)
	if err != nil {
		return errors.Wrapf(err, "opening db connection")
	}
	defer db.Close()
	// filpathWalkDir does not evaluate the PFS symlink
	sym, err := filepath.EvalSymlinks(inputFolder)
	if err != nil {
		return errors.Wrapf(err, "evaluating symlinks")
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
	logPerf(ctx, "job info done")
	if err != nil {
		return errors.Wrapf(err, "walking file tree for job info ")
	}
	if len(jobInfoPaths) == 0 {
		log.Info(ctx, "No new job info found, exiting without inserting test results.")
		return nil
	}
	for _, path := range testResultPaths {
		if err = insertTestResultFile(ctx, path, jobInfoPaths, db); err != nil {
			return errors.Wrapf(err, "inserting test results into DB ")
		}
	}
	logPerf(ctx, "test results done")
	return nil
}

func readJobInfo(path string) (*gotestresults.JobInfo, error) {
	jobInfoFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "opening job info json file")
	}
	defer jobInfoFile.Close()
	jobInfoJson, err := io.ReadAll(jobInfoFile)
	if err != nil {
		return nil, errors.Wrapf(err, "reading job info data")
	}
	var jobInfo gotestresults.JobInfo
	if err = json.Unmarshal(jobInfoJson, &jobInfo); err != nil {
		return nil, errors.Wrapf(err, "unmarshalling job info json")
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
							pull_requests,
							username
						) VALUES (
							:workflowid, 
							:jobid, 
							:jobexecutor, 
							:jobname, 
							:jobtimestamp, 
							:jobnumexecutors, 
							:commit, 
							:branch, 
							:tag, 
							:pullrequests, 
							:username)`,
		jobInfo,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Unique key constraint violation. In other words, the record already exists.
			// Ignore insert to since we don't want to change the timestamp, but log it.
			log.Info(ctx, "Job info already inserted, Skipping insert for workflow/job",
				zap.String("Workflow id", jobInfo.WorkflowId),
				zap.String("Job id", jobInfo.JobId),
				zap.Int("Job Executor", jobInfo.JobExecutor),
			)
			return nil
		}
		return errors.Wrapf(err, "inserting job info for workflow %v  job %v  executor %v",
			jobInfo.WorkflowId,
			jobInfo.JobId,
			jobInfo.JobExecutor,
		)
	}
	jobInfoPaths[strings.TrimSuffix(path, d.Name())] = *jobInfo // for use when inserting individual test results, don't insert results if errored
	log.Info(ctx, "Inserted job info for workflow and job",
		zap.String("Workflow id", jobInfo.WorkflowId),
		zap.String("Job id", jobInfo.JobId),
		zap.Int("Job Executor", jobInfo.JobExecutor),
	)
	return nil
}

func insertTestResultFile(ctx context.Context, path string, jobInfoPaths map[string]gotestresults.JobInfo, db *sqlx.DB) error {
	resultsFile, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "opening results file ")
	}
	defer resultsFile.Close()
	// get job info previously inserted in order to link foreign keys to jobs table
	fileName := filepath.Base(path)
	jobInfo, ok := jobInfoPaths[strings.TrimSuffix(path, fileName)]
	if !ok {
		return errors.WithStack(fmt.Errorf(
			"Failed to find job info for %v - file name %v - job infos %v ",
			path,
			filepath.Base(path),
			jobInfoPaths,
		))
	}
	query := `INSERT INTO public.test_results (
		id,
		workflow_id,
		job_id,
		job_executor,
		test_name,
		package,
		"action",
		elapsed_seconds,
		"output"
	) VALUES (:id, :workflowid, :jobid, :jobexecutor, :testname, :package, :action, :elapsedseconds, :output)`
	paramVals := &[]ResultRow{}
	resultsScanner := bufio.NewScanner(resultsFile)
	egUpload, ctx := errgroup.WithContext(ctx)
	chunkIdx := 0
	egUpload.SetLimit(uploadGoroutineLimit)
	for resultsScanner.Scan() {
		resultJson := resultsScanner.Bytes()
		var result TestResultLine
		err = json.Unmarshal(resultJson, &result)
		if err != nil {
			log.Info(ctx, "Failed unmarshalling file as a test result. Continuing to parse results file.",
				zap.String("path", path),
				zap.Error(err),
			)
			continue
		}
		if result.Action == "output" {
			continue // output is mostly logs and coverage and it's not that useful, so we don't store output-only actions
		}
		*paramVals = append(*paramVals, ResultRow{
			Id:             uuid.New().String(),
			WorkflowId:     jobInfo.WorkflowId,
			JobId:          jobInfo.JobId,
			JobExecutor:    jobInfo.JobExecutor,
			TestName:       result.TestName,
			Package:        result.Package,
			Action:         result.Action,
			ElapsedSeconds: result.Elapsed,
			Output:         result.Output,
		})
		chunkIdx++
		if chunkIdx >= uploadChunkSize {
			chunkIdx = 0
			flushResults(ctx, db, egUpload, query, paramVals)
			paramVals = &[]ResultRow{}
		}
	}
	if chunkIdx > 0 { // we have a partial chunk that needs to upload
		flushResults(ctx, db, egUpload, query, paramVals)
	}
	err = egUpload.Wait()
	if err != nil {
		return errors.EnsureStack(err)
	}
	if err = resultsScanner.Err(); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

func flushResults(ctx context.Context, db *sqlx.DB, egUpload *errgroup.Group, query string, paramVals *[]ResultRow) {
	threadParamVals := paramVals
	egUpload.Go(func() error { // upload in goroutine so scanner can continue
		if _, err := db.NamedExec(query, *threadParamVals); err != nil {
			log.Info(ctx, "Failed updating SQL DB - continuing to parse results file.", zap.Error(err))
		}
		return nil
	})
}

func logPerf(ctx context.Context, msg string) {
	log.Info(ctx, fmt.Sprintf("Performance: %s", msg), zap.Duration("time since start", time.Since(startTime)))
}
