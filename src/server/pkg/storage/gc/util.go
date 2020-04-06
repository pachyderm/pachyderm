package gc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

func NewLocalServer(deleter Deleter) (Client, error) {
	server, err := NewServer(deleter, "localhost", 32228, nil)
	if err != nil {
		return nil, err
	}
	return NewClient(server, "localhost", 32228, nil)
}

func WithLocalServer(deleter Deleter, f func(Client) error) error {
	client, err := NewLocalServer(deleter)
	if err != nil {
		return err
	}
	return f(client)
}

// TODO: connection options
func openDatabase(host string, port uint16) (*gorm.DB, error) {
	return gorm.Open("postgres", fmt.Sprintf("host=%s port=%d dbname=pgc user=pachyderm password=elephantastic sslmode=disable", host, port))
}

func readChunksFromCursor(cursor *sql.Rows) []string {
	chunks := []string{}
	for cursor.Next() {
		var hash string
		if err := cursor.Scan(&hash); err != nil {
			return nil
		}
		chunks = append(chunks, hash)
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
func reserveChunkStats(err error, start time.Time) {
	applySQLStats("reserveChunk", err, start)
}

type statementFunc func(*gorm.DB) *gorm.DB

func runTransaction(ctx context.Context, db *gorm.DB, stmtFuncs []statementFunc, statsCallback func(error, time.Time)) error {
	for {
		//start := time.Now()
		err := tryTransaction(ctx, db, stmtFuncs)
		//statsCallback(err, start)
		if err == nil {
			return nil
		} else if !isRetriableError(err) {
			return err
		}
	}
}

func tryTransaction(ctx context.Context, db *gorm.DB, stmtFuncs []statementFunc) error {
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

	for _, stmtFunc := range stmtFuncs {
		if err := stmtFunc(txn).Error; err != nil {
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
