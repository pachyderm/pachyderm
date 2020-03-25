package gc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

func WithLocalServer(f func(c Client) error) error {
	s := makeServer()
}

// TODO: connection options
func openDatabase(host string, port uint16) (*gorm.DB, error) {
	return gorm.Open("postgres", fmt.Sprintf("host=%s port=%d dbname=pgc user=pachyderm password=elephantastic sslmode=disable", host, port))
}

func readChunksFromCursor(cursor *sql.Rows) []chunk.Chunk {
	chunks := []chunk.Chunk{}
	for cursor.Next() {
		var hash string
		if err := cursor.Scan(&hash); err != nil {
			return nil
		}
		chunks = append(chunks, chunk.Chunk{Hash: hash})
	}
	return chunks
}

func isRetriableError(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		return err.Code.Class().Name() == "transaction_rollback"
	}
	return false
}

// stats callbacks for use with runTransaction
func markChunksDeletingStats(err error, start time.Time) {
	applySQLStats("markChunksDeleting", err, start)
}
func removeChunkRowsStats(err error, start time.Time) {
	applySQLStats("removeChunkRows", err, start)
}
func reserveChunksStats(err error, start time.Time) {
	applySQLStats("reserveChunks", err, start)
}
func updateReferencesStats(err error, start time.Time) {
	applySQLStats("updateReferences", err, start)
}

type stmtCallback func(*gorm.DB) *gorm.DB

func runTransaction(
	ctx context.Context,
	db *gorm.DB,
	statements []stmtCallback,
	statsCallback func(error, time.Time),
) error {
	for {
		start := time.Now()
		err := tryTransaction(ctx, db, statements)
		statsCallback(err, start)
		if err == nil {
			return nil
		} else if !isRetriableError(err) {
			return err
		}
	}
}

func tryTransaction(
	ctx context.Context,
	db *gorm.DB,
	statements []stmtCallback,
) error {
	txn := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})

	// So we don't leak in case of a panic somewhere unexpected
	defer func() {
		if r := recover(); r != nil {
			txn.Rollback()
		}
	}()

	if err := txn.Error; err != nil {
		return err
	}

	for _, stmt := range statements {
		if err := stmt(txn).Error; err != nil {
			txn.Rollback()
			return err
		}
	}

	if err := txn.Commit().Error; err != nil {
		txn.Rollback()
		return err
	}

	return nil
}
