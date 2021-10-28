package transforms

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/sdata"
	"github.com/pachyderm/pachyderm/v2/src/internal/secrets"
	"github.com/sirupsen/logrus"
)

type SQLIngestParams struct {
	// Instrumentation
	Logger *logrus.Logger

	// PFS In/Out
	InputDir, OutputDir string

	URL      pachsql.URL
	Password secrets.Secret
	Query    string
	Format   string
	// Shard affects the number appended to the file path.
	Shard int
}

// SQLIngest connects to a SQL database at params.URL and runs params.Query.
// The resulting rows are written to files in params.OutputDir.
// The format of the output file is controlled by params.Format.
// Valid options are "json" and "csv"
//
// It makes outgoing connections using pachsql.OpenURL
// It accesses the filesystem only within params.InputDir, and params.OutputDir
func SQLIngest(ctx context.Context, params SQLIngestParams) error {
	log := params.Logger
	log.Infof("Connecting to DB at %v...", params.URL)
	db, err := pachsql.OpenURL(params.URL, string(params.Password))
	if err != nil {
		return err
	}
	defer db.Close()
	if err := func() error {
		ctx, cf := context.WithTimeout(ctx, 10*time.Second)
		defer cf()
		return db.PingContext(ctx)
	}(); err != nil {
		return err
	}
	log.Infof("Connected to DB")
	writerFactory, err := makeWriterFactory(params.Format)
	if err != nil {
		return err
	}
	outputPath := filepath.Join(params.OutputDir, fmt.Sprintf("%04d", params.Shard))
	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	log.Infof("Writing output to %q", outputPath)
	log.Infof("Begin query %q", params.Query)
	rows, err := db.QueryContext(ctx, params.Query)
	if err != nil {
		return err
	}
	log.Infof("Query complete, begin reading rows...")
	colNames, err := rows.Columns()
	if err != nil {
		return err
	}
	log.Infof("Column names: %v", colNames)
	tw := writerFactory(outputFile, colNames)
	res, err := sdata.MaterializeSQL(tw, rows)
	if err != nil {
		return err
	}
	log.Infof("Successfully materialized %d rows", res.RowCount)
	if err := outputFile.Close(); err != nil {
		return err
	}
	log.Infof("DONE")
	return nil
}

type writerFactory = func(w io.Writer, fieldNames []string) sdata.TupleWriter

func makeWriterFactory(formatName string) (writerFactory, error) {
	var factory writerFactory
	switch formatName {
	case "json", "jsonlines":
		factory = func(w io.Writer, fieldNames []string) sdata.TupleWriter {
			return sdata.NewJSONWriter(w, fieldNames)
		}
	case "csv":
		factory = func(w io.Writer, fieldNames []string) sdata.TupleWriter {
			return sdata.NewCSVWriter(w, fieldNames)
		}
	default:
		return nil, errors.Errorf("unrecognized format %v", formatName)
	}
	return factory, nil
}
