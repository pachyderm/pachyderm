package persist

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/dancannon/gorethink"
)

type migrationFunc func(address string, databaseName string) error

// MissingMigrationErr denotes that no migration is supported for the provided versions
type MissingMigrationErr struct {
	error
}

func newMissingMigrationErr(msg error) MissingMigrationErr {
	return MissingMigrationErr{msg}
}

var (
	migrationMap = map[string]migrationFunc{
		"1.3.4-1.3.6": oneThreeFourToOneThreeSix,
	}
)

// Migrate updates the database schema only in the forward direction
func Migrate(address, databaseName, migrationKey string) error {
	migrate, ok := migrationMap[migrationKey]
	if !ok {
		return newMissingMigrationErr(fmt.Errorf("migration %s is not supported for %v", migrationKey, databaseName))
	}
	return migrate(address, databaseName)
}

// 1.3.4 -> 1.3.6
func oneThreeFourToOneThreeSix(address, databaseName string) error {
	session, err := DbConnect(address)
	if err != nil {
		return err
	}
	log.Infof("Renaming Diff 'Size' to 'SizeBytes'")
	if _, err := gorethink.DB(databaseName).Table(diffTable).Replace(
		func(row gorethink.Term) gorethink.Term {
			return row.Without("Size").Merge(
				map[string]gorethink.Term{
					"SizeBytes": row.Field("Size"),
				},
			)
		},
	).RunWrite(session); err != nil {
		return err
	}
	log.Infof("Renaming Repo 'Size' to 'SizeBytes'")
	if _, err := gorethink.DB(databaseName).Table(repoTable).Replace(
		func(row gorethink.Term) gorethink.Term {
			return row.Without("Size").Merge(
				map[string]gorethink.Term{
					"SizeBytes": row.Field("Size"),
				},
			)
		},
	).RunWrite(session); err != nil {
		return err
	}
	log.Infof("Migration succeeded")
	return nil
}
