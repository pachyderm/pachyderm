package main

import (
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/secrets"
	"github.com/pachyderm/pachyderm/v2/src/internal/transforms"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	pfsIn  = "/pfs/in"
	pfsOut = "/pfs/out"
)

func main() {
	ctx := context.Background()
	log := logrus.StandardLogger()
	if len(os.Args) < 1 {
		log.Fatal("1 argument required")
	}
	transformName := os.Args[0]
	transformArgs := os.Args[1:]
	switch transformName {
	case "sql-ingest":
		if err := sqlIngest(ctx, log, transformArgs); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("unrecognized transform name %q", transformName)
	}
}

func sqlIngest(ctx context.Context, log *logrus.Logger, args []string) error {
	if len(args) < 3 {
		return errors.Errorf("must provide db url, format, and query")
	}
	urlStr, formatName, query := args[0], args[1], args[2]
	const passwordEnvar = "PACHYDERM_SQL_PASSWORD"
	password, ok := os.LookupEnv(passwordEnvar)
	if !ok {
		return errors.Errorf("must set %v", passwordEnvar)
	}
	u, err := pachsql.ParseURL(urlStr)
	if err != nil {
		return err
	}
	return transforms.SQLIngest(ctx, transforms.SQLIngestParams{
		Logger: log,

		InputDir:  pfsIn,
		OutputDir: pfsOut,

		URL:      *u,
		Password: secrets.Secret(password),
		Query:    query,
		Format:   formatName,
	})
}
