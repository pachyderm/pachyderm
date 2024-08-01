package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/sdata"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func getEgressPassword() (string, error) {
	const passwordEnvar = "PACHYDERM_SQL_PASSWORD" // TODO move this to a package for sharing
	password, ok := os.LookupEnv(passwordEnvar)
	if !ok {
		return "", errors.Errorf("must set %v", passwordEnvar)
	}
	return password, nil
}

func (d *driver) copyToObjectStorage(ctx context.Context, taskService task.Service, file *pfs.File, destURL string) (*pfs.EgressResponse_ObjectStorageResult, error) {
	bytesWritten, err := d.getFileURL(ctx, taskService, destURL, file, nil)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	result := new(pfs.EgressResponse_ObjectStorageResult)
	result.BytesWritten = bytesWritten
	return result, nil
}

func copyToSQLDB(ctx context.Context, src Source, destURL string, fileFormat *pfs.SQLDatabaseEgress_FileFormat) (*pfs.EgressResponse_SQLDatabaseResult, error) {
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
	result := new(pfs.EgressResponse_SQLDatabaseResult)
	result.RowsWritten = make(map[string]int64)
	err = src.Iterate(ctx, func(fi *pfs.FileInfo, file fileset.File) error {
		if fi.FileType != pfs.FileType_FILE {
			return nil
		}

		tableName := strings.Split(fi.File.Path, "/")[1]
		tableInfo, ok := tableInfos[tableName]
		if !ok {
			// first time interacting with table, so do a full drop first
			// TODO figure out how to full sync better
			tableInfo, err = pachsql.GetTableInfoTx(tx, tableName)
			if err != nil {
				return errors.EnsureStack(err)
			}
			tableInfos[tableName] = tableInfo
			_, err := tx.Exec(fmt.Sprintf("DELETE FROM %s.%s", tableInfo.Schema, tableInfo.Name))
			if err != nil {
				return errors.EnsureStack(err)
			}
		}

		if err := miscutil.WithPipe(
			func(w io.Writer) error {
				return errors.EnsureStack(file.Content(ctx, w))
			},
			func(r io.Reader) error {
				var tr sdata.TupleReader
				switch fileFormat.Type {
				case pfs.SQLDatabaseEgress_FileFormat_CSV:
					tr = sdata.NewCSVParser(r).WithHeaderFields(fileFormat.Columns)
				case pfs.SQLDatabaseEgress_FileFormat_JSON:
					tr = sdata.NewJSONParser(r, fileFormat.Columns)
				default:
					return errors.Errorf("unknown file format %v", fileFormat.Type)
				}
				tw := sdata.NewSQLTupleWriter(tx, tableInfo)
				tuple, err := sdata.NewTupleFromTableInfo(tableInfo)
				if err != nil {
					return errors.EnsureStack(err)
				}
				n, err := sdata.Copy(tw, tr, tuple)
				result.RowsWritten[tableName] += int64(n)
				return errors.EnsureStack(err)
			}); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return result, errors.EnsureStack(tx.Commit())
}
