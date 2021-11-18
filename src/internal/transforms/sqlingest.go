package transforms

import (
	"context"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/sdata"
	"github.com/pachyderm/pachyderm/v2/src/internal/secrets"
	"github.com/sirupsen/logrus"
)

// SQLIngestParams are the parameters passed to SQLIngest
type SQLIngestParams struct {
	// Instrumentation
	Logger *logrus.Logger

	// PFS In/Out
	InputDir, OutputDir string

	URL      pachsql.URL
	Password secrets.Secret
	Format   string
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
	if err := bijectiveMap(params.InputDir, params.OutputDir, IdentityPM, func(r io.Reader, w io.Writer) error {
		queryBytes, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		query := string(queryBytes)
		log.Infof("Query: %q", query)
		log.Info("Running query...")
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		log.Infof("Query complete, begin reading rows...")
		colNames, err := rows.Columns()
		if err != nil {
			return err
		}
		log.Infof("Column names: %v", colNames)
		tw := writerFactory(w, colNames)
		res, err := sdata.MaterializeSQL(tw, rows)
		if err != nil {
			return err
		}
		log.Infof("Successfully materialized %d rows", res.RowCount)
		return nil
	}); err != nil {
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
