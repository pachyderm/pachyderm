package persist

import (
	log "github.com/Sirupsen/logrus"
	"github.com/dancannon/gorethink"
)

type migrationFunc func(address string, databaseName string) error

var (
	migrationMap = map[string]migrationFunc{
		"1.3.4-1.3.6": oneThreeFourToOneThreeSix,
	}
)

// Migrate updates the database schema only in the forward direction
func Migrate(address, databaseName, migrationKey string) error {
	migrate, ok := migrationMap[migrationKey]
	if !ok {
		return fmt.Errorf("migration %s is not supported", migrationKey)
	}
	return migrate(address, databaseName)
}

// 1.3.4 -> 1.3.6
func oneThreeFourToOneThreeSix(address, databaseName string) error {
	session, err := connect(address)
	if err != nil {
		return err
	}
	log.Infof("Renaming 'Size' to 'SizeBytes'")
	if _, err := gorethink.DB(databaseName).Table(diffsTable).Replace(
		func(row gorethink.Term) {
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
