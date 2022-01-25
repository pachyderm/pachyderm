package main

import (
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/secrets"
	"github.com/pachyderm/pachyderm/v2/src/internal/transforms"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	pfs    = "/pfs"
	pfsOut = "/pfs/out"
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

var entrypoints = map[string]Entrypoint{
	"sql-ingest":      sqlIngest,
	"sql-gen-queries": sqlGenQueries,
}

func sqlIngest(ctx context.Context, log *logrus.Logger, args []string) error {
	const passwordEnvar = "PACHYDERM_SQL_PASSWORD"
	if len(args) < 2 {
		return errors.Errorf("must provide db url and format")
	}
	urlStr, formatName := args[0], args[1]
	password, ok := os.LookupEnv(passwordEnvar)
	if !ok {
		return errors.Errorf("must set %v", passwordEnvar)
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

		URL:      *u,
		Password: secrets.Secret(password),
		Format:   formatName,
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
