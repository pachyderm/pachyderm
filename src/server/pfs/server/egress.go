package server

import (
	"io"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/sdata"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"golang.org/x/net/context"
)

type egressResult map[string]int64

func getEgressPassword() (string, error) {
	const passwordEnvar = "PACHYDERM_SQL_PASSWORD" // TODO move this to a package for sharing
	password, ok := os.LookupEnv(passwordEnvar)
	if !ok {
		return "", errors.Errorf("must set %v", passwordEnvar)
	}
	return password, nil
}

func copyToSQLDB(ctx context.Context, src Source, destURL string, fileFormat *pfs.SQLEgressOptions_FileFormat) (egressResult, error) {
	url, err := pachsql.ParseURL(destURL)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	// do we need to wrap password in a secret?
	password, err := getEgressPassword()
	if err != nil {
		return nil, err
	}
	db, err := pachsql.OpenURL(*url, password)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer db.Close()

	// all table are written through a single transaction
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer tx.Rollback()

	// cache tableInfos because multiple files can belong to the same table
	tableInfos := make(map[string]*pachsql.TableInfo)
	rowsWritten := make(egressResult)
	err = src.Iterate(ctx, func(fi *pfs.FileInfo, file fileset.File) error {
		if fi.FileType != pfs.FileType_FILE {
			return nil
		}

		tableName := strings.Split(fi.File.Path, "/")[1]
		tableInfo, prs := tableInfos[tableName]
		if !prs {
			tableInfo, err = pachsql.GetTableInfoTx(tx, tableName)
			if err != nil {
				return errors.EnsureStack(err)
			}
			tableInfos[tableName] = tableInfo
		}

		if err := miscutil.WithPipe(
			func(w io.Writer) error {
				return errors.EnsureStack(file.Content(ctx, w))
			},
			func(r io.Reader) error {
				var tr sdata.TupleReader
				switch fileFormat.Type {
				case pfs.SQLEgressOptions_FileFormat_CSV:
					tr = sdata.NewCSVParser(r)
				case pfs.SQLEgressOptions_FileFormat_JSON:
					tr = sdata.NewJSONParser(r, fileFormat.Options.JsonFieldNames)
				}
				tw := sdata.NewSQLTupleWriter(tx, tableInfo)
				tuple, err := sdata.NewTupleFromTableInfo(tableInfo)
				if err != nil {
					return errors.EnsureStack(err)
				}
				n, err := sdata.Copy(tr, tw, tuple)
				rowsWritten[tableName] += int64(n)
				return errors.EnsureStack(err)
			}); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return rowsWritten, errors.EnsureStack(tx.Commit())
}
