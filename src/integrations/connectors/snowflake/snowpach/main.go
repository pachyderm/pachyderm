package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	log "github.com/sirupsen/logrus"
	sf "github.com/snowflakedb/gosnowflake"
)

var (
	header, debug                               bool
	query, partitionBy, fileFormat, copyOptions string
	inputDir, outputDir, targetTable            string
)

func env(k string, failOnMissing bool) string {
	if value := os.Getenv(k); value != "" {
		return value
	}
	log.Fatalf("missing environment variable %v", k)
	return ""
}

// SNOWSQL_X env variables are defined by the snowsql CLI
// we inherit this convention for convenience
func getDSN() (string, error) {
	account := env("SNOWSQL_ACCOUNT", true)
	user := env("SNOWSQL_USER", true)
	role := env("SNOWSQL_ROLE", false)
	warehouse := env("SNOWSQL_WH", false)
	database := env("SNOWSQL_DATABASE", true)
	schema := env("SNOWSQL_SCHEMA", false)
	password := env("SNOWSQL_PWD", true)

	cfg := &sf.Config{
		Account:   account,
		User:      user,
		Role:      role,
		Warehouse: warehouse,
		Database:  database,
		Schema:    schema,
		Password:  password,
	}

	return sf.DSN(cfg) //nolint:wrapcheck
}

func createTempStage(db *sqlx.DB, stage string) error {
	// Create a named Snowflake stage based off of the pipeline name
	// should this be a temporary stage?
	if result, err := db.Exec(fmt.Sprintf("CREATE TEMPORARY STAGE %s", stage)); err != nil {
		fmt.Println(result)
		return errors.Errorf("error creating temp stage: %s, error: %v", stage, err)
	}
	return nil
}

func copyIntoStage(db *sqlx.DB, stage string) error {
	if err := createTempStage(db, stage); err != nil {
		return err
	}
	// run COPY INTO <stage> FROM <query>
	copyIntoQuery := fmt.Sprintf("COPY INTO @%s FROM (%s)", stage, query)
	if partitionBy != "" {
		copyIntoQuery += fmt.Sprintf(" PARTITION BY (%s)", partitionBy)
	}
	if fileFormat != "" {
		copyIntoQuery += fmt.Sprintf(" FILE_FORMAT = (%s)", fileFormat)
	}
	if copyOptions != "" {
		copyIntoQuery += fmt.Sprintf(" %s", copyOptions)
	}
	if header {
		copyIntoQuery += " HEADER = TRUE"
	}
	if _, err := db.Exec(copyIntoQuery); err != nil {
		return errors.Errorf("error copying data to stage: %v", err)
	}
	return nil
}

func listFromStage(db *sqlx.DB, stage string) ([]string, error) {
	// to list files of a named stage: list @<stage>;
	rows, err := db.Query(fmt.Sprintf("list @%s", stage))
	if err != nil {
		return nil, errors.Errorf("error list @%s: %v", stage, err)
	}
	defer rows.Close()
	var files []string
	for rows.Next() {
		var (
			name, size, md5, last_modified string
		)
		if err = rows.Scan(&name, &size, &md5, &last_modified); err != nil {
			return nil, errors.Errorf("error scanning results of list stage: %v", err)
		}
		files = append(files, name)
	}
	return files, nil
}

func downloadFromStage(db *sqlx.DB, stage string, files []string, outputDir string) error {
	// todo make this run goroutines for each file
	// download files from stage to /pfs/out
	// file shards can be of the path pipelineName/a/b/c/filename.txt
	// we want to write this to /pfs/out/a/b/c/filename.txt so we need to create the parent directoires
	for _, file := range files {
		outputFile := filepath.Join(outputDir, strings.TrimPrefix(file, stage))
		outputFileDir := filepath.Dir(outputFile)
		if err := os.MkdirAll(outputFileDir, 0775); err != nil {
			return errors.Errorf("could not make parent dir %s: %v", outputFileDir, err)
		}
		if _, err := db.Exec(fmt.Sprintf("GET @%s file://%s", file, outputFileDir)); err != nil {
			return errors.Errorf("could not download @%s to %s: %v", file, outputFileDir, err)
		}
		log.Infof("downloaded %s to %s", file, outputFileDir)
	}
	return nil
}

func connect() (*sqlx.DB, error) {
	dsn, err := getDSN()
	if err != nil {
		log.Fatalf("failed to get Snowflake DSN from environment: %s, error: %v", dsn, err)
	}
	db, err := sqlx.Open("snowflake", dsn)
	if err != nil {
		return nil, errors.Errorf("failed to connect to db: %v, error: %v", db, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Errorf("error verifying connection to database: %v", err)
	}
	return db, nil
}

func read(db *sqlx.DB) error {
	stage := env("PPS_PIPELINE_NAME", true) + env("PACH_JOB_ID", false)
	if err := copyIntoStage(db, stage); err != nil {
		return err
	}

	files, err := listFromStage(db, stage)
	if err != nil {
		return err
	}

	if err := downloadFromStage(db, stage, files, outputDir); err != nil {
		return err
	}
	return nil
}

func writeTable(db *sqlx.DB, files []string, table string) error {
	if len(files) == 0 {
		return nil
	}
	// create temporary stage to stage files
	stage := env("PPS_PIPELINE_NAME", true) + env("PACH_JOB_ID", false)
	if err := createTempStage(db, stage); err != nil {
		return err
	}

	// for all files in /pfs/in, run PUT file:///pfs/in/<filepath> @%<table>
	for _, file := range files {
		in := "file://" + file
		out := fmt.Sprintf("@%s/%s", stage, table) + filepath.Dir(file)
		if _, err := db.Exec(fmt.Sprintf("PUT %s %s", in, out)); err != nil {
			return errors.Errorf("failed to upload file to stage: %v", err)
		}
	}

	// in a single transaction run DELETE FROM table; COPY INTO <table> FROM <table_stage>
	ctx := context.Background()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
		return errors.Errorf("failed to clear table: %v", err)
	}
	copyIntoQuery := fmt.Sprintf("COPY INTO %s FROM @%s/%s", table, stage, table)
	if fileFormat != "" {
		copyIntoQuery += fmt.Sprintf(" FILE_FORMAT = (%s)", fileFormat)
	}
	if copyOptions != "" {
		copyIntoQuery += fmt.Sprintf(" %s", copyOptions)
	}

	if _, err := tx.ExecContext(ctx, copyIntoQuery); err != nil {
		return errors.Errorf("failed to load data from stage into table: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}

// Two modes
// mode 1: no target table is provided, so infer table names based on top level directory
// mode 2: target table name is provided, so all files in input directory load into that table
func write(db *sqlx.DB) error {
	switch targetTable {
	case "":
		// TODO
		// infer target table based on /pfs/in/tablename
	default:
		// remember that /pfs/in has symbolic links so filepath.Walk() doesn't work
		files, err := filepath.Glob(filepath.Join(inputDir, "/*/*"))
		if err != nil {
			return errors.Errorf("failed to list files for writing to database: %v", err)
		}
		return writeTable(db, files, targetTable)
	}
	return nil
}

func main() {
	// TODO should we use cobra?
	readCmd := flag.NewFlagSet("read", flag.ExitOnError)
	readCmd.BoolVar(&debug, "debug", false, "set Snowflake log level to 'debug'")
	readCmd.StringVar(&query, "query", "", "a SQL query")
	readCmd.StringVar(&partitionBy, "partitionBy", "", "expression for the partition key")
	readCmd.StringVar(&fileFormat, "fileFormat", "", "configure file format options")
	readCmd.StringVar(&copyOptions, "copyOptions", "", "configure options for copying files to stage")
	readCmd.StringVar(&outputDir, "outputDir", "", "local directory to save results of query")
	readCmd.BoolVar(&header, "header", false, "whether to include header in the output files")

	writeCmd := flag.NewFlagSet("write", flag.ExitOnError)
	writeCmd.BoolVar(&debug, "debug", false, "set Snowflake log level to 'debug'")
	writeCmd.StringVar(&inputDir, "inputDir", "", "example /pfs/in")
	writeCmd.StringVar(&targetTable, "table", "", "target table")
	writeCmd.StringVar(&fileFormat, "fileFormat", "", "configure file format options")
	writeCmd.StringVar(&copyOptions, "copyOptions", "", "configure options for copying files to stage")
	if len(os.Args) < 2 {
		log.Fatal("expected 'read' or 'write' subcommands")
		os.Exit(1)
	}

	// attempt to connect to database and verify connection
	db, err := connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// snowflake logger can tell us what queries are running, useful for debugging
	sfLogger := sf.GetLogger()
	switch os.Args[1] {
	case "read":
		if err := readCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		if debug {
			_ = sfLogger.SetLogLevel("debug")
		}
		if err := read(db); err != nil {
			log.Fatal(err)
		}
	case "write":
		if err := writeCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		if debug {
			_ = sfLogger.SetLogLevel("debug")
		}
		if err := write(db); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("subcommand must be either 'read' or 'write'")
		os.Exit(1)
	}
}
