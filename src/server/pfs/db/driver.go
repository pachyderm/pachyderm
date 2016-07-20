package persist

import (
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	libclock "github.com/pachyderm/pachyderm/src/server/pfs/db/clock"
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"

	"github.com/dancannon/gorethink"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/time"
	"google.golang.org/grpc"
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
	diffTable   Table = "Diffs"
	clockTable  Table = "Clocks"

	commitTable Table = "Commits"
	// commitBranchIndex maps commits to branches
	commitBranchIndex Index = "CommitBranchIndex"

	connectTimeoutSeconds = 5
)

var (
	tables = []Table{
		repoTable,
		branchTable,
		commitTable,
		diffTable,
		clockTable,
	}

	tableToTableCreateOpts = map[Table][]gorethink.TableCreateOpts{
		repoTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: "Name",
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
		clockTable: []gorethink.TableCreateOpts{
			gorethink.TableCreateOpts{
				PrimaryKey: "ID",
			},
		},
	}
)

type driver struct {
	blockAddress string
	blockClient  pfs.BlockAPIClient

	dbAddress string
	dbName    string
	dbClient  *gorethink.Session
}

func NewDriver(blockAddress string, dbAddress string, dbName string) (drive.Driver, error) {
	clientConn, err := grpc.Dial(blockAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	dbClient, err := dbConnect(dbAddress)
	if err != nil {
		return nil, err
	}

	return &driver{
		blockAddress: blockAddress,
		blockClient:  pfs.NewBlockAPIClient(clientConn),
		dbAddress:    dbAddress,
		dbName:       dbName,
		dbClient:     dbClient,
	}, nil
}

func InitDB(address string, databaseName string) error {
	session, err := dbConnect(address)
	if err != nil {
		return err
	}
	defer session.Close()

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
	if _, err := gorethink.DB(databaseName).Table(commitTable).IndexCreateFunc(commitBranchIndex, func(row gorethink.Term) interface{} {
		return row.Field("BranchClocks").Map(func(branchClock gorethink.Term) interface{} {
			lastClock := branchClock.Field("Clocks").Nth(-1)
			return []interface{}{
				lastClock.Field("Branch"),
				lastClock.Field("Clock"),
			}
		})
	}, gorethink.IndexCreateOpts{
		Multi: true,
	}).RunWrite(session); err != nil {
		return err
	}

	// Wait for indexes to be ready
	if _, err := gorethink.DB(databaseName).Table(commitTable).IndexWait(commitBranchIndex).RunWrite(session); err != nil {
		return err
	}

	return nil
}

func RemoveDB(address string, databaseName string) error {
	session, err := dbConnect(address)
	if err != nil {
		return err
	}
	defer session.Close()

	// Create the database
	if _, err := gorethink.DBDrop(databaseName).RunWrite(session); err != nil {
		return err
	}

	return nil
}

func dbConnect(address string) (*gorethink.Session, error) {
	return gorethink.Connect(gorethink.ConnectOpts{
		Address: address,
		Timeout: connectTimeoutSeconds * time.Second,
	})
}

func clockToID(c *persist.Clock) *persist.ClockID {
	return &persist.ClockID{fmt.Sprintf("%s/%d", c.Branch, c.Clock)}
}

func validateRepoName(name string) error {
	match, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", name)

	if !match {
		return fmt.Errorf("repo name (%v) invalid: only alphanumeric and underscore characters allowed", name)
	}

	return nil
}

func (d *driver) getTerm(table Table) gorethink.Term {
	return gorethink.DB(d.dbName).Table(table)
}

func (d *driver) CreateRepo(repo *pfs.Repo, created *google_protobuf.Timestamp,
	provenance []*pfs.Repo, shards map[uint64]bool) error {

	err := validateRepoName(repo.Name)
	if err != nil {
		return err
	}

	_, err = d.getTerm(repoTable).Insert(&persist.Repo{
		Name:    repo.Name,
		Created: created,
	}).RunWrite(d.dbClient)
	return err
}

func (d *driver) InspectRepo(repo *pfs.Repo, shards map[uint64]bool) (repoInfo *pfs.RepoInfo, retErr error) {
	cursor, err := d.getTerm(repoTable).Get(repo.Name).Run(d.dbClient)
	if err != nil {
		return nil, err
	}
	rawRepo := &persist.Repo{}
	if err := cursor.One(rawRepo); err != nil {
		return nil, err
	}
	repoInfo = &pfs.RepoInfo{
		Repo:    &pfs.Repo{rawRepo.Name},
		Created: rawRepo.Created,
	}
	return repoInfo, nil
}

func (d *driver) ListRepo(provenance []*pfs.Repo, shards map[uint64]bool) (repoInfos []*pfs.RepoInfo, retErr error) {
	cursor, err := d.getTerm(repoTable).Run(d.dbClient)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	for {
		repo := &persist.Repo{}
		if !cursor.Next(repo) {
			break
		}
		repoInfos = append(repoInfos, &pfs.RepoInfo{
			Repo:    &pfs.Repo{repo.Name},
			Created: repo.Created,
		})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return repoInfos, nil
}

func (d *driver) DeleteRepo(repo *pfs.Repo, shards map[uint64]bool, force bool) error {
	_, err := d.getTerm(repoTable).Get(repo.Name).Delete().RunWrite(d.dbClient)
	return err
}

func (d *driver) StartCommit(repo *pfs.Repo, commitID string, parentID string, branch string, started *google_protobuf.Timestamp, provenance []*pfs.Commit, shards map[uint64]bool) (retErr error) {
	var _provenance []string
	for _, c := range provenance {
		_provenance = append(_provenance, c.ID)
	}
	commit := &persist.Commit{
		ID:         commitID,
		Repo:       repo.Name,
		Started:    prototime.TimeToTimestamp(time.Now()),
		Provenance: _provenance,
	}

	if parentID == "" {
		if branch == "" {
			branch = uuid.NewWithoutDashes()
		}

		_, err := d.getTerm(branchTable).Insert(&persist.Branch{
			ID: branchID(repo.Name, branch),
		}, gorethink.InsertOpts{
			// Set conflict to "replace" because we don't want to get an error
			// when the branch already exists.
			Conflict: "replace",
		}).RunWrite(d.dbClient)
		if err != nil {
			return err
		}

		for {
			cursor, err := d.getTerm(commitTable).OrderBy(gorethink.OrderByOpts{
				Index: gorethink.Desc(commitBranchIndex),
			}).Between(
				[]interface{}{branch, 0},
				[]interface{}{branch, gorethink.MaxVal},
			).Run(d.dbClient)
			if err != nil {
				return err
			}

			// The last commit on this branch will be our parent commit
			parentCommit := &persist.Commit{}
			err = cursor.One(parentCommit)
			if err != nil && err != gorethink.ErrEmptyResult {
				return err
			} else if err == gorethink.ErrEmptyResult {
				// we don't have a parent :(
				// so we create a new BranchClock
				commit.BranchClocks = libclock.NewBranchClocks(branch)
			} else {
				fmt.Println("got a parent!")
				// we do have a parent :D
				// so we inherit our parent's branch clock for this particular branch,
				// and increment the last component by 1
				commit.BranchClocks, err = libclock.NewChildOfBranchClocks(parentCommit.BranchClocks, branch)
				if err != nil {
					return err
				}
			}
			clock, err := libclock.GetClockForBranch(commit.BranchClocks, branch)
			if err != nil {
				return err
			}
			clockID := clockToID(clock)
			fmt.Printf("branch clocks: %v\n", commit.BranchClocks)
			_, err = d.getTerm(clockTable).Insert(clockID, gorethink.InsertOpts{
				// we want to return an error in case the clock already exists
				Conflict: "error",
			}).RunWrite(d.dbClient)
			if gorethink.IsConflictErr(err) {
				// Try again with a new clock
				fmt.Println("retrying")
				continue
			} else if err != nil {
				return err
			}
			break
		}
	}

	_, err := d.getTerm(commitTable).Insert(commit).RunWrite(d.dbClient)

	return err
}

func branchID(repo string, name string) string {
	return fmt.Sprintf("%s/%s", repo, name)
}

// FinishCommit blocks until its parent has been finished/cancelled
func (d *driver) FinishCommit(commit *pfs.Commit, finished *google_protobuf.Timestamp, cancel bool, shards map[uint64]bool) error {
	_, err := d.getTerm(commitTable).Get(commit.ID).Update(
		map[string]interface{}{
			"Finished": finished,
		},
	).RunWrite(d.dbClient)

	return err
}

func (d *driver) InspectCommit(commit *pfs.Commit, shards map[uint64]bool) (commitInfo *pfs.CommitInfo, retErr error) {
	cursor, err := d.getTerm(commitTable).Get(commit.ID).Run(d.dbClient)
	if err != nil {
		return nil, err
	}

	rawCommit := &persist.Commit{}
	if err := cursor.One(rawCommit); err != nil {
		return nil, err
	}

	return rawCommitToCommitInfo(rawCommit), nil
}

func rawCommitToCommitInfo(rawCommit *persist.Commit) *pfs.CommitInfo {
	commitType := pfs.CommitType_COMMIT_TYPE_READ
	if rawCommit.Finished == nil {
		commitType = pfs.CommitType_COMMIT_TYPE_WRITE
	}
	return &pfs.CommitInfo{
		Commit: &pfs.Commit{
			Repo: &pfs.Repo{rawCommit.Repo},
			ID:   rawCommit.ID,
		},
		Started:    rawCommit.Started,
		Finished:   rawCommit.Finished,
		CommitType: commitType,
	}
}

func (d *driver) ListCommit(repos []*pfs.Repo, commitType pfs.CommitType, fromCommit []*pfs.Commit,
	provenance []*pfs.Commit, all bool, shards map[uint64]bool) (commitInfos []*pfs.CommitInfo, retErr error) {
	cursor, err := d.getTerm(commitTable).Filter(func(commit gorethink.Term) gorethink.Term {
		var predicates []interface{}
		for _, repo := range repos {
			predicates = append(predicates, commit.Field("Repo").Eq(repo.Name))
		}
		return gorethink.Or(predicates...)
	}).Run(d.dbClient)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	for {
		rawCommit := &persist.Commit{}
		if !cursor.Next(rawCommit) {
			break
		}
		commitInfos = append(commitInfos, rawCommitToCommitInfo(rawCommit))
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return commitInfos, nil
}

func (d *driver) ListBranch(repo *pfs.Repo, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	return nil, nil
}

func (d *driver) DeleteCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	return nil
}

func (d *driver) PutFile(file *pfs.File, handle string,
	delimiter pfs.Delimiter, shard uint64, reader io.Reader) (retErr error) {
	return nil
}

func (d *driver) MakeDirectory(file *pfs.File, shard uint64) (retErr error) {
	return nil
}

func (d *driver) GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64,
	size int64, from *pfs.Commit, shard uint64, unsafe bool, handle string) (io.ReadCloser, error) {
	return nil, nil
}

func (d *driver) InspectFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64, unsafe bool, handle string) (*pfs.FileInfo, error) {
	return nil, nil
}

func (d *driver) ListFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64, recurse bool, unsafe bool, handle string) ([]*pfs.FileInfo, error) {
	return nil, nil
}

func (d *driver) DeleteFile(file *pfs.File, shard uint64, unsafe bool, handle string) error {
	return nil
}

func (d *driver) DeleteAll(shards map[uint64]bool) error {
	return nil
}

func (d *driver) AddShard(shard uint64) error {
	return nil
}

func (d *driver) DeleteShard(shard uint64) error {
	return nil
}

func (d *driver) Dump() {
}
