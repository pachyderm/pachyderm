package persist

import (
	"time"

	"github.com/dancannon/gorethink"
)

// A Table is a rethinkdb table name.
type Table string

// A PrimaryKey is a rethinkdb primary key identifier.
type PrimaryKey string

// An Index is a rethinkdb index.
type Index string

const (
	repoTable   Table = "Repos"
	branchTable Table = "Branches"

	commitTable       Table = "Commits"
	commitClocksIndex Index = "CommitClocksIndex"
	commitRepoIndex   Index = "CommitRepoIndex"

	diffTable         Table = "Diffs"
	diffCommitIDIndex Index = "DiffCommitIDIndex"
	diffPathIndex     Index = "DiffPathIndex"
)

var (
	tables = []Table{
		repoTable,
		branchTable,
		commitTable,
		diffTable,
	}

	tableToTableCreateOpts = map[Table][]gorethink.TableCreateOpts{
		repoTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: "name",
			},
		},
		branchTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: "ID",
			},
		},
		commitTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: "ID",
			},
		},
		diffTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: "ID",
			},
		},
	}
)

func InitDB(address string, databaseName string) error {
	session, err := connect(address)
	if err != nil {
		return err
	}

	// Create the database
	if _, err := gorethink.DBCreate(databaseName).RunWrite(session); err != nil {
		return err
	}

	// Create tables
	for _, table := range tables {
		tableCreateOpts := tableToTableCreateOpts[table]
		if _, err := gorethink.DB(databaseName).TableCreate(table, tableCreateOpts...).RunWrite(session); err != nil {
			return err
		}
	}

	// Create indexes
	if _, err := gorethink.DB(databaseName).Table(commitTable).IndexCreate(commitClocksIndex).RunWrite(session); err != nil {
		return err
	}
	if _, err := gorethink.DB(databaseName).Table(commitTable).IndexCreate(commitRepoIndex).RunWrite(session); err != nil {
		return err
	}
	if _, err := gorethink.DB(databaseName).Table(diffTable).IndexCreate(diffCommitIDIndex).RunWrite(session); err != nil {
		return err
	}
	if _, err := gorethink.DB(databaseName).Table(diffTable).IndexCreate(diffPathIndex).RunWrite(session); err != nil {
		return err
	}

	return nil
}

func connect(address string) (*gorethink.Session, error) {
	return gorethink.Connect(gorethink.ConnectOpts{
		Address: address,
		Timeout: connectTimeoutSeconds * time.Second,
	})
}
