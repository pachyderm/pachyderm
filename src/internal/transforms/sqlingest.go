package transforms

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

// SQLIngest connects to a SQL database at params.URL and runs queries
// read from files in the input.
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
		return errors.EnsureStack(db.PingContext(ctx))
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
			return errors.EnsureStack(err)
		}
		query := string(queryBytes)
		log.Infof("Query: %q", query)
		log.Info("Running query...")
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return errors.EnsureStack(err)
		}
		log.Infof("Query complete, begin reading rows...")
		colNames, err := rows.Columns()
		if err != nil {
			return errors.EnsureStack(err)
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
			return sdata.NewCSVWriter(w, nil)
		}
	default:
		return nil, errors.Errorf("unrecognized format %v", formatName)
	}
	return factory, nil
}

type SQLQueryGenerationParams struct {
	Logger              *logrus.Logger
	InputDir, OutputDir string

	Query string
}

// SQLQueryGeneration generates queries with a timestamp in the comments
func SQLQueryGeneration(ctx context.Context, params SQLQueryGenerationParams) error {
	timestamp, err := readCronTimestamp(params.Logger, params.InputDir)
	if err != nil {
		return err
	}
	timestampComment := fmt.Sprintf("-- %d\n", timestamp)
	contents := timestampComment + params.Query
	outputPath := filepath.Join(params.OutputDir, "0000")
	return errors.EnsureStack(ioutil.WriteFile(outputPath, []byte(contents), 0755))
}

func readCronTimestamp(log *logrus.Logger, inputDir string) (uint64, error) {
	dirEnts, err := os.ReadDir(inputDir)
	if err != nil {
		return 0, errors.EnsureStack(err)
	}
	for _, dirEnt := range dirEnts {
		name := dirEnt.Name()
		timestamp, err := time.Parse(time.RFC3339, name)
		if err != nil {
			log.Errorf("could not parse %q into timestamp", name)
			continue
		}
		log.Infof("found cron timestamp %q", name)
		timestamp = timestamp.UTC()
		return uint64(timestamp.Unix()), nil
	}
	return 0, errors.Errorf("missing timestamp file")
}
