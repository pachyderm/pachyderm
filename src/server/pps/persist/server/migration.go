package server

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pps"

	log "github.com/Sirupsen/logrus"
	"github.com/dancannon/gorethink"
	"github.com/gogo/protobuf/types"
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
		"1.2.4-1.3.0": oneTwoFourToOneThreeZero,
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

// 1.2.4 -> 1.3.0
func oneTwoFourToOneThreeZero(address, databaseName string) error {
	session, err := connect(address)
	if err != nil {
		return err
	}
	log.Infof("Creating %s index", pipelineNameAndStateAndGCIndex)
	if _, err := gorethink.DB(databaseName).Table(jobInfosTable).IndexCreateFunc(
		pipelineNameAndStateAndGCIndex,
		func(row gorethink.Term) interface{} {
			return []interface{}{
				row.Field("PipelineName"),
				row.Field("State"),
				row.Field("Gc"),
			}
		}).RunWrite(session); err != nil {
		return err
	}

	log.Infof("Adding default GC policy to all pipelines")
	if _, err := gorethink.DB(databaseName).Table(pipelineInfosTable).Update(map[string]interface{}{
		"GcPolicy": &pps.GCPolicy{
			Success: &types.Duration{
				Seconds: 24 * 60 * 60,
			},
			Failure: &types.Duration{
				Seconds: 7 * 24 * 60 * 60,
			},
		},
	}).Run(session); err != nil {
		return err
	}

	log.Infof("Adding GC flag to all jobs")
	if _, err := gorethink.DB(databaseName).Table(jobInfosTable).Update(map[string]interface{}{
		"Gc": false,
	}).Run(session); err != nil {
		return err
	}

	log.Infof("Migration succeeded")
	return nil
}
