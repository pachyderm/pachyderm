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
	switch transformName {
	case "sql-ingest":
		if err := sqlIngest(ctx, log, transformArgs); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unrecognized transform name %q", transformName)
	}
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
	inputDir := filepath.FromSlash(pfs + "/in")
	outputDir := filepath.FromSlash(pfsOut)
	return transforms.SQLIngest(ctx, transforms.SQLIngestParams{
		Logger: log,

		InputDir:  inputDir,
		OutputDir: outputDir,

		URL:      *u,
		Password: secrets.Secret(password),
		Format:   formatName,
	})
}
