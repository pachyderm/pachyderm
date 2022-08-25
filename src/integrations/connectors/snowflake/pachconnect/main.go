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
	"github.com/sirupsen/logrus"
	sf "github.com/snowflakedb/gosnowflake"
)

var (
	header, debug                                          bool
	query, outputDir, partitionBy, fileFormat, copyOptions string
)

func init() {
	flag.StringVar(&query, "query", "", "a SQL query")
	flag.StringVar(&partitionBy, "partitionBy", "", "expression for the partition key")
	flag.StringVar(&fileFormat, "fileFormat", "", "configure file format options")
	flag.StringVar(&copyOptions, "copyOptions", "", "configure options for copying files to stage")
	flag.StringVar(&outputDir, "outputDir", "/pfs/out", "local directory to save results of query")
	flag.BoolVar(&header, "header", false, "whether to include header in the output files")
	flag.BoolVar(&debug, "debug", false, "set to true for logging debug messages from Snowflake")
}

var (
	log = logrus.StandardLogger()
)

func env(k string, failOnMissing bool) string {
	if value := os.Getenv(k); value != "" {
		return value
	}
	log.Fatalf("missing environment variable %v", k)
	return ""
}

func getDSN() (string, error) {
	account := env("SNOWSQL_ACCOUNT", true)
	user := env("SNOWSQL_USER", true)
	role := env("SNOWSQL_ROLE", false)
	warehouse := env("SNOWSQL_WH", false)
	database := env("SNOWSQL_DATABASE", true)
	schema := env("SNOWSQL_SCHEMA", false)
	password := env("SNOWSQL_PASSWORD", true)

	cfg := &sf.Config{
		Account:   account,
		User:      user,
		Role:      role,
		Warehouse: warehouse,
		Database:  database,
		Schema:    schema,
		Password:  password,
	}

	return sf.DSN(cfg)
}

func copyIntoStage(db *sqlx.DB, stage string) error {
	// Create a named Snowflake stage based off of the pipeline name
	// should this be a temporary stage?
	if _, err := db.Exec(fmt.Sprintf("CREATE OR REPLACE TEMPORARY STAGE %s", stage)); err != nil {
		return errors.Errorf("error creating named stage: %s, error: %v", stage, err)
	}

	// run COPY INTO <stage> FROM <query>
	copyIntoQuery := fmt.Sprintf("COPY INTO @%s FROM (%s)", stage, query)
	if partitionBy != "" {
		copyIntoQuery += fmt.Sprintf(" PARTITION BY = %s", partitionBy)
	}
	if fileFormat != "" {
		copyIntoQuery += fmt.Sprintf(" FILE_FORMAT = %s", fileFormat)
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
	// download files from stage to /pfs/out
	// file shards can be of the path pipelineName/a/b/c/filename.txt
	// we want to write this to /pfs/out/a/b/c/filename.txt so we need to create the parent directoires
	for _, file := range files {
		outputFile := filepath.Join(outputDir, strings.TrimPrefix(file, stage))
		outputFileDir, _ := filepath.Split(outputFile)
		if err := os.MkdirAll(outputFileDir, 0777); err != nil {
			return errors.Errorf("could not make parent dir %s: %v", outputFileDir, err)
		}
		if _, err := db.Exec(fmt.Sprintf("GET @%s file://%s", file, outputFileDir)); err != nil {
			return errors.Errorf("could not download @%s to %s: %v", file, outputFileDir, err)
		}
		log.Infof("downloaded %s", file)
	}
	return nil
}

func main() {
	flag.Parse()

	sfLogger := sf.GetLogger()
	if debug {
		sfLogger.SetLogLevel("debug")
	}

	dsn, err := getDSN()
	if err != nil {
		log.Fatalf("failed to get Snowflake DSN from environment: %s, error: %v", dsn, err)
	}

	db, err := sqlx.Open("snowflake", dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v, error: %v", db, err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("error verifying connection to database: %v", err)
	}

	stage := env("PPS_PIPELINE_NAME", true)
	if err := copyIntoStage(db, stage); err != nil {
		log.Fatal(err)
	}

	files, err := listFromStage(db, stage)
	if err != nil {
		log.Fatal(err)
	}

	// todo make this run goroutines for each file
	if err := downloadFromStage(db, stage, files, outputDir); err != nil {
		log.Fatal(err)
	}
}
