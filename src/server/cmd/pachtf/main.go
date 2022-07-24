package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/breml/rootcerts"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/secrets"
	"github.com/pachyderm/pachyderm/v2/src/internal/transforms"
	"github.com/sirupsen/logrus"
)

const (
	pfs                    = "/pfs"
	pfsOut                 = "/pfs/out"
	PACHYDERM_SQL_PASSWORD = "PACHYDERM_SQL_PASSWORD"
)

func main() {
	ctx := context.Background()
	log := logrus.StandardLogger()
	args := os.Args[1:]
	if len(args) < 1 {
		log.Fatal("at least 1 argument required")
	}
	transformName := args[0]
	transformArgs := args[1:]
	entrypoint, ok := entrypoints[transformName]
	if !ok {
		log.Fatalf("unrecognized transform name %q", transformName)
	}
	ents, err := os.ReadDir(filepath.FromSlash(pfs))
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Listing /pfs")
	for _, ent := range ents {
		log.Info("/pfs/" + ent.Name())
	}
	if err := entrypoint(ctx, log, transformArgs); err != nil {
		log.Fatal(err)
	}
}

type Entrypoint = func(ctx context.Context, log *logrus.Logger, args []string) error

// sql-gen-queries and sql-ingest are deprecated, and users should prefer sql-run instead.
var entrypoints = map[string]Entrypoint{
	"sql-ingest":      sqlIngest,
	"sql-gen-queries": sqlGenQueries,
	"sql-run":         sqlRun,
}

func sqlIngest(ctx context.Context, log *logrus.Logger, args []string) error {
	if len(args) < 2 {
		return errors.Errorf("must provide db url and format")
	}
	urlStr, formatName := args[0], args[1]
	var hasHeader bool
	if len(args) > 2 {
		var err error
		hasHeader, err = strconv.ParseBool(args[2])
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
	password, ok := os.LookupEnv(PACHYDERM_SQL_PASSWORD)
	if !ok {
		return errors.Errorf("must set %v", PACHYDERM_SQL_PASSWORD)
	}
	u, err := pachsql.ParseURL(urlStr)
	if err != nil {
		return err
	}
	log.Infof("DB protocol=%v host=%v port=%v database=%v\n", u.Protocol, u.Host, u.Port, u.Database)
	inputDir, err := filepath.EvalSymlinks(filepath.FromSlash(pfs + "/in"))
	if err != nil {
		return errors.EnsureStack(err)
	}
	outputDir, err := filepath.EvalSymlinks(filepath.FromSlash(pfsOut))
	if err != nil {
		return errors.EnsureStack(err)
	}
	return transforms.SQLIngest(ctx, transforms.SQLIngestParams{
		Logger: log,

		InputDir:  inputDir,
		OutputDir: outputDir,

		URL:       *u,
		Password:  secrets.Secret(password),
		Format:    formatName,
		HasHeader: hasHeader,
	})
}

func sqlGenQueries(ctx context.Context, log *logrus.Logger, args []string) error {
	if len(args) < 1 {
		return errors.Errorf("must provide query")
	}
	query := args[0]
	inputDir, err := filepath.EvalSymlinks(filepath.FromSlash(pfs + "/in"))
	if err != nil {
		return errors.EnsureStack(err)
	}
	outputDir, err := filepath.EvalSymlinks(filepath.FromSlash(pfsOut))
	if err != nil {
		return errors.EnsureStack(err)
	}
	return transforms.SQLQueryGeneration(ctx, transforms.SQLQueryGenerationParams{
		Logger:    log,
		InputDir:  inputDir,
		OutputDir: outputDir,
		Query:     query,
	})
}

func sqlRun(ctx context.Context, log *logrus.Logger, args []string) error {
	if len(args) < 5 {
		return errors.Errorf("must provide url fileFormat query outputFile and hasHeader")
	}
	url, fileFormat, query, oFname := args[0], args[1], args[2], args[3]
	hasHeader, err := strconv.ParseBool(args[4])
	if err != nil {
		return errors.EnsureStack(err)
	}

	password, ok := os.LookupEnv(PACHYDERM_SQL_PASSWORD)
	if !ok {
		return errors.Errorf("must set %v", PACHYDERM_SQL_PASSWORD)
	}
	u, err := pachsql.ParseURL(url)
	if err != nil {
		return err
	}

	params := transforms.SQLRunParams{
		Logger:     log,
		OutputDir:  pfsOut,
		OutputFile: oFname,
		Query:      query,
		Password:   secrets.Secret(password),
		URL:        *u,
		HasHeader:  hasHeader,
		Format:     fileFormat,
	}
	return transforms.RunSQLRaw(ctx, params)
}
