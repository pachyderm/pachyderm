package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/breml/rootcerts"
	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/secrets"
	"github.com/pachyderm/pachyderm/v2/src/internal/transforms"
)

const (
	pfs                    = "/pfs"
	pfsOut                 = "/pfs/out"
	PACHYDERM_SQL_PASSWORD = "PACHYDERM_SQL_PASSWORD"
)

func main() {
	log.InitPachctlLogger()
	ctx := pctx.Background("")
	args := os.Args[1:]
	if len(args) < 1 {
		log.Error(ctx, "at least 1 argument required")
		os.Exit(1)
	}
	transformName := args[0]
	transformArgs := args[1:]
	entrypoint, ok := entrypoints[transformName]
	if !ok {
		log.Error(ctx, "unrecognized transform name", zap.String("name", transformName))
		os.Exit(1)
	}
	ents, err := os.ReadDir(filepath.FromSlash(pfs))
	if err != nil {
		log.Error(ctx, "problem reading directory", zap.Error(err))
		os.Exit(1)
	}
	log.Info(ctx, "Listing /pfs")
	for _, ent := range ents {
		log.Info(ctx, "/pfs/"+ent.Name())
	}
	if err := entrypoint(ctx, transformArgs); err != nil {
		log.Error(ctx, "problem running job", zap.Error(err))
		os.Exit(1)
	}
}

type Entrypoint = func(ctx context.Context, args []string) error

// sql-gen-queries and sql-ingest are deprecated, and users should prefer sql-run instead.
var entrypoints = map[string]Entrypoint{
	"sql-ingest":      sqlIngest,
	"sql-gen-queries": sqlGenQueries,
	"sql-run":         sqlRun,
}

func sqlIngest(ctx context.Context, args []string) error {
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
	log.Info(ctx, "ingesting from database", zap.Stringer("database", u))
	inputDir, err := filepath.EvalSymlinks(filepath.FromSlash(pfs + "/in"))
	if err != nil {
		return errors.EnsureStack(err)
	}
	outputDir, err := filepath.EvalSymlinks(filepath.FromSlash(pfsOut))
	if err != nil {
		return errors.EnsureStack(err)
	}
	return transforms.SQLIngest(ctx, transforms.SQLIngestParams{
		InputDir:  inputDir,
		OutputDir: outputDir,
		URL:       *u,
		Password:  secrets.Secret(password),
		Format:    formatName,
		HasHeader: hasHeader,
	})
}

func sqlGenQueries(ctx context.Context, args []string) error {
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
		InputDir:  inputDir,
		OutputDir: outputDir,
		Query:     query,
	})
}

func sqlRun(ctx context.Context, args []string) error {
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