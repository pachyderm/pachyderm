package transforms

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	glob "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/sdata"
	"github.com/pachyderm/pachyderm/v2/src/internal/secrets"
	"github.com/sirupsen/logrus"
)

type SQLEgressParams struct {
	InputDir string

	URL        pachsql.URL
	Password   secrets.Secret
	Format     string
	TableGlobs map[string]string

	Logger *logrus.Logger
}

func SQLEgress(ctx context.Context, params SQLEgressParams) error {
	log := params.Logger
	db, err := pachsql.OpenURL(params.URL, string(params.Password))
	if err != nil {
		return err
	}
	defer db.Close()
	var readerFactory func([]string, io.Reader) sdata.TupleReader
	switch params.Format {
	case "json":
		readerFactory = func(colNames []string, r io.Reader) sdata.TupleReader {
			return sdata.NewJSONParser(r, colNames)
		}
	case "csv":
		readerFactory = func(colNames []string, r io.Reader) sdata.TupleReader {
			return sdata.NewCSVParser(r)
		}
	default:
		return errors.Errorf("format not supported: %q", params.Format)
	}
	var count int64
	if err := dbutil.WithTx(ctx, db, func(tx *pachsql.Tx) error {
		for tableName, globStr := range params.TableGlobs {
			g, err := glob.Compile(globStr)
			if err != nil {
				return err
			}
			tableInfo, err := pachsql.LookupTableInfo(tx, tableName)
			if err != nil {
				return err
			}
			// create the loader for the table
			tw := sdata.NewSQLTableLoader(tx, pachsql.MakeInsertStatement(tx.DriverName(), tableName, tableInfo.ColumnNames))
			if err := filepath.WalkDir(params.InputDir, func(name string, ent fs.DirEntry, err error) error {
				if !g.Match(name) {
					return nil
				}
				// open the file
				var f *os.File
				tr := readerFactory(tableInfo.ColumnNames, f)
				row := sdata.NewTupleFromSQL(tableInfo.ColumnTypes)
				n, err := sdata.Copy(tw, tr, row)
				if err != nil {
					return err
				}
				count += int64(n)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	log.Info("successfully egressed %d rows to %d tables", count, len(params.TableGlobs))
	return nil
}
