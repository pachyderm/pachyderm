package persist

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	libclock "github.com/pachyderm/pachyderm/src/server/pfs/db/clock"
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"

	"github.com/dancannon/gorethink"
	"github.com/gogo/protobuf/proto"
	"go.pedge.io/lion/proto"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/time"
	"google.golang.org/grpc"
)

// A Table is a rethinkdb table name.
type Table string

// A PrimaryKey is a rethinkdb primary key identifier.
type PrimaryKey string

// Errors
type ErrCommitNotFound struct {
	error
}

type ErrBranchExists struct {
	error
}

const (
	repoTable   Table = "Repos"
	branchTable Table = "Branches"
	diffTable   Table = "Diffs"
	clockTable  Table = "Clocks"
	commitTable Table = "Commits"

	connectTimeoutSeconds = 5
)

const (
	// this is used in the "appends" field to signify deletion
	DeleteAppend = "delete"
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
	blockClient pfs.BlockAPIClient
	dbName      string
	dbClient    *gorethink.Session
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
		blockClient: pfs.NewBlockAPIClient(clientConn),
		dbName:      dbName,
		dbClient:    dbClient,
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
	for _, someIndex := range Indexes {
		if _, err := gorethink.DB(databaseName).Table(someIndex.GetTable()).IndexCreateFunc(someIndex.GetName(), someIndex.GetCreateFunction(), someIndex.GetCreateOptions()).RunWrite(session); err != nil {
			return err
		}
		if _, err := gorethink.DB(databaseName).Table(someIndex.GetTable()).IndexWait(someIndex.GetName()).RunWrite(session); err != nil {
			return err
		}
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

	if branch == "" {
		branch = uuid.NewWithoutDashes()
	}

	var clockID *persist.ClockID
	if parentID == "" {
		for {
			// The head of this branch will be our parent commit
			parentCommit := &persist.Commit{}
			err := d.getHeadOfBranch(repo.Name, branch, parentCommit)
			if err != nil && err != gorethink.ErrEmptyResult {
				return err
			} else if err == gorethink.ErrEmptyResult {
				// we don't have a parent :(
				// so we create a new BranchClock
				commit.BranchClocks = libclock.NewBranchClocks(branch)
			} else {
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
			clockID = getClockID(repo.Name, clock)
			err = d.insertMessage(clockTable, clockID)
			if gorethink.IsConflictErr(err) {
				// There is another process creating a commit on this branch
				// at the same time.  We lost the race, but we can try again
				continue
			} else if err != nil {
				return err
			}
			break
		}
	} else {
		parentCommit := &persist.Commit{}
		parentClock, err := parseClock(parentID)
		if err == nil {
			if err := d.getMessageByIndex(commitTable, CommitBranchIndex, CommitBranchIndex.Key(repo.Name, parentClock.Branch, parentClock.Clock), parentCommit); err != nil {
				return err
			}
		} else {
			// OBSOLETE
			// This logic is here to make the implementation compatible with the
			// old API where parentID can be a UUID, instead of a semantically
			// meaningful ID such as "master/0".
			// This logic should be removed eventually once we migrate to the
			// new API.
			if err := d.getMessageByPrimaryKey(commitTable, parentID, parentCommit); err != nil {
				return err
			}
		}

		var parentBranch string
		if parentClock == nil {
			// OBSOLETE
			parentBranch = parentCommit.BranchClocks[0].Clocks[len(parentCommit.BranchClocks[0].Clocks)-1].Branch
		} else {
			parentBranch = parentClock.Branch
		}

		commit.BranchClocks, err = libclock.NewBranchOffBranchClocks(parentCommit.BranchClocks, parentBranch, branch)
		if err != nil {
			return err
		}

		clock, err := libclock.GetClockForBranch(commit.BranchClocks, branch)
		if err != nil {
			return err
		}

		clockID = getClockID(repo.Name, clock)
		if err := d.insertMessage(clockTable, clockID); err != nil {
			if gorethink.IsConflictErr(err) {
				// This should only happen if there's another process creating the
				// very same branch at the same time, and we lost the race.
				return ErrBranchExists{fmt.Errorf("branch %s already exists", branch)}
			}
			return err
		}
	}
	defer func() {
		if retErr != nil {
			if err := d.deleteMessageByPrimaryKey(clockTable, clockID.ID); err != nil {
				protolion.Debugf("Unable to remove clock after StartCommit fails; this will result in database inconsistency")
			}
		}
	}()

	// TODO: what if the program exits here?  There will be an entry in the Clocks
	// table, but not in the Commits table.  Now you won't be able to create this
	// commit anymore.
	return d.insertMessage(commitTable, commit)
}

func (d *driver) getHeadOfBranch(repo string, branch string, commit *persist.Commit) error {
	cursor, err := d.betweenIndex(
		commitTable, CommitBranchIndex.GetName(),
		CommitBranchIndex.Key(repo, branch, 0),
		CommitBranchIndex.Key(repo, branch, gorethink.MaxVal),
		true,
	).Run(d.dbClient)
	if err != nil {
		return err
	}
	return cursor.One(commit)
}

func getClockID(repo string, c *persist.Clock) *persist.ClockID {
	return &persist.ClockID{
		ID:     fmt.Sprintf("%s/%s/%d", repo, c.Branch, c.Clock),
		Repo:   repo,
		Branch: c.Branch,
		Clock:  c.Clock,
	}
}

// parseClock takes a string of the form "branch/clock"
// and returns a Clock object.
// For example:
// "master/0" -> Clock{"master", 0}
func parseClock(clock string) (*persist.Clock, error) {
	parts := strings.Split(clock, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid commit ID %s")
	}
	c, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid commit ID %s")
	}
	return &persist.Clock{
		Branch: parts[0],
		Clock:  uint64(c),
	}, nil
}

type CommitChangeFeed struct {
	NewVal *persist.Commit `gorethink:"new_val,omitempty"`
}

// FinishCommit blocks until its parent has been finished/cancelled
func (d *driver) FinishCommit(commit *pfs.Commit, finished *google_protobuf.Timestamp, cancel bool, shards map[uint64]bool) error {
	// OBSOLETE
	// Once we move to the new implementation, we should be able to directly
	// Infer the parentID using the given commit ID.
	parentID, err := d.getIDOfParentCommit(commit.Repo.Name, commit.ID)
	if err != nil {
		return err
	}

	rawCommit, err := d.getCommitByAmbiguousID(commit.Repo.Name, commit.ID)
	if err != nil {
		return err
	}

	// TODO: validate diffs

	var parentCancelled bool
	if parentID != "" {
		cursor, err := d.getTerm(commitTable).Get(parentID).Changes(gorethink.ChangesOpts{
			IncludeInitial: true,
		}).Run(d.dbClient)

		if err != nil {
			return err
		}

		var change CommitChangeFeed
		for cursor.Next(&change) {
			if change.NewVal != nil && change.NewVal.Finished != nil {
				parentCancelled = change.NewVal.Cancelled
				break
			}
		}
		if err = cursor.Err(); err != nil {
			return err
		}
	}

	rawCommit.Finished = finished
	rawCommit.Cancelled = parentCancelled || cancel
	_, err = d.getTerm(commitTable).Get(rawCommit.ID).Update(rawCommit).RunWrite(d.dbClient)
	return err
}

func (d *driver) InspectCommit(commit *pfs.Commit, shards map[uint64]bool) (commitInfo *pfs.CommitInfo, retErr error) {
	rawCommit, err := d.getCommitByAmbiguousID(commit.Repo.Name, commit.ID)
	if err != nil {
		return nil, err
	}
	commitInfo = rawCommitToCommitInfo(rawCommit)
	// OBSOLETE
	// Old API Server expects request commit ID to match results commit ID
	commitInfo.Commit.ID = commit.ID
	return commitInfo, nil
}

func rawCommitToCommitInfo(rawCommit *persist.Commit) *pfs.CommitInfo {
	commitType := pfs.CommitType_COMMIT_TYPE_READ
	var branch string
	if len(rawCommit.BranchClocks) > 0 {
		branch = libclock.GetBranchNameFromBranchClock(rawCommit.BranchClocks[0])
	}
	if rawCommit.Finished == nil {
		commitType = pfs.CommitType_COMMIT_TYPE_WRITE
	}
	return &pfs.CommitInfo{
		Commit: &pfs.Commit{
			Repo: &pfs.Repo{rawCommit.Repo},
			ID:   rawCommit.ID,
		},
		Branch:     branch,
		Started:    rawCommit.Started,
		Finished:   rawCommit.Finished,
		Cancelled:  rawCommit.Cancelled,
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
	// Get all branches
	cursor, err := d.getTerm(clockTable).Between(
		[]interface{}{repo.Name, gorethink.MinVal},
		[]interface{}{repo.Name, gorethink.MaxVal},
		gorethink.BetweenOpts{
			Index: ClockBranchIndex.GetName(),
		},
	).Distinct().Field("Branch").Run(d.dbClient)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var branches []string
	if err := cursor.All(&branches); err != nil {
		return nil, err
	}

	// OBSOLETE
	// To maintain API compatibility, we return the heads of the branches
	var commitInfos []*pfs.CommitInfo
	for _, branch := range branches {
		commit := &persist.Commit{}
		if err := d.getHeadOfBranch(repo.Name, branch, commit); err != nil {
			return nil, err
		}
		commitInfos = append(commitInfos, &pfs.CommitInfo{
			Commit: &pfs.Commit{
				Repo: repo,
				ID:   commit.ID,
			},
			Branch: branch,
		})
	}
	return commitInfos, nil
}

func (d *driver) DeleteCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	return nil
}

func (d *driver) PutFile(file *pfs.File, handle string,
	delimiter pfs.Delimiter, shard uint64, reader io.Reader) (retErr error) {
	// TODO: eventually optimize this with a cache so that we don't have to
	// go to the database to figure out if the commit exists
	commit, err := d.getCommitByAmbiguousID(file.Commit.Repo.Name, file.Commit.ID)
	if err != nil {
		return err
	}

	_client := client.APIClient{BlockAPIClient: d.blockClient}
	blockrefs, err := _client.PutBlock(delimiter, reader)
	if err != nil {
		return err
	}

	var refs []*persist.BlockRef
	var size uint64
	for _, blockref := range blockrefs.BlockRef {
		ref := &persist.BlockRef{
			Hash:  blockref.Block.Hash,
			Upper: blockref.Range.Upper,
			Lower: blockref.Range.Lower,
		}
		refs = append(refs, ref)
		size += ref.Size()
	}

	_, err = d.getTerm(diffTable).Insert(&persist.Diff{
		ID:           getDiffID(commit.ID, file.Path),
		Repo:         commit.Repo,
		Delete:       false,
		CommitID:     commit.ID,
		Path:         file.Path,
		BlockRefs:    refs,
		Size:         size,
		BranchClocks: commit.BranchClocks,
	}, gorethink.InsertOpts{
		Conflict: func(id gorethink.Term, oldDoc gorethink.Term, newDoc gorethink.Term) gorethink.Term {
			return oldDoc.Merge(map[string]interface{}{
				"BlockRefs": oldDoc.Field("BlockRefs").Add(newDoc.Field("BlockRefs")),
				"Size":      oldDoc.Field("Size").Add(newDoc.Field("Size")),
				"Delete":    false,
			})
		},
	}).RunWrite(d.dbClient)
	return err
}

func getDiffID(commitID string, path string) string {
	return fmt.Sprintf("%s:%s", commitID, path)
}

// the equivalent of above except that commitID is a rethink term
func getDiffIDFromTerm(commitID gorethink.Term, path string) gorethink.Term {
	return commitID.Add(":" + path)
}

func (d *driver) MakeDirectory(file *pfs.File, shard uint64) (retErr error) {
	return nil
}

func reverseSlice(s [][]*persist.BranchClock) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func (d *driver) GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64,
	size int64, from *pfs.Commit, shard uint64, unsafe bool, handle string) (io.ReadCloser, error) {
	var fromClocks *persist.BranchClock
	if from != nil {
		fromCommit, err := d.getCommitByAmbiguousID(from.Repo.Name, from.ID)
		if err != nil {
			return nil, err
		}
		fromClocks = fromCommit.BranchClocks[0]
	}

	toCommit, err := d.getCommitByAmbiguousID(file.Commit.Repo.Name, file.Commit.ID)
	if err != nil {
		return nil, err
	}
	toClocks := toCommit.BranchClocks[0]

	// Find the most recent diff that removes the path
	intervals, err := libclock.GetClockIntervals(fromClocks, toClocks)
	if err != nil {
		return nil, err
	}
	// Reverse because we want to find the most recent diff
	reverseSlice(intervals)

	cursor, err := gorethink.Expr(intervals).ConcatMap(func(interval gorethink.Term) gorethink.Term {
		firstClock := persist.BranchClockToArray(interval.Nth(0))
		secondClock := persist.BranchClockToArray(interval.Nth(1))
		return gorethink.Range(gorethink.Expr(0).Sub(secondClock.Nth(-1).Nth(1)), gorethink.Expr(0).Sub(firstClock.Nth(-1).Nth(1)).Add(1)).Map(func(x gorethink.Term) gorethink.Term {
			return firstClock.ChangeAt(-1, firstClock.Nth(-1).ChangeAt(1, gorethink.Expr(0).Sub(x)))
		})
	}).Map(func(clocks gorethink.Term) interface{} {
		return DiffPathIndex.Key(file.Commit.Repo.Name, true, file.Path, clocks)
	}).EqJoin(func(x gorethink.Term) gorethink.Term {
		// no-op
		return x
	}, d.getTerm(diffTable), gorethink.EqJoinOpts{
		Index:   DiffPathIndex.GetName(),
		Ordered: true,
	}).Field("right").Run(d.dbClient)
	if err != nil {
		return nil, err
	}

	diff := &persist.Diff{}
	err = cursor.One(diff)
	if err != nil && err != gorethink.ErrEmptyResult {
		return nil, err
	}

	// Stream from the most recent delete to the current commit
	if err != gorethink.ErrEmptyResult {
		intervals, err = libclock.GetClockIntervals(diff.BranchClocks[0], toClocks)
		if err != nil {
			return nil, err
		}
	} else {
		intervals, err = libclock.GetClockIntervals(fromClocks, toClocks)
		if err != nil {
			return nil, err
		}
	}

	cursor, err = gorethink.Expr(intervals).ConcatMap(func(interval gorethink.Term) gorethink.Term {
		firstClock := persist.BranchClockToArray(interval.Nth(0))
		secondClock := persist.BranchClockToArray(interval.Nth(1))
		return gorethink.Range(firstClock.Nth(-1).Nth(1), secondClock.Nth(-1).Nth(1).Add(1)).Map(func(x gorethink.Term) gorethink.Term {
			return firstClock.ChangeAt(-1, firstClock.Nth(-1).ChangeAt(1, x))
		}).Map(func(clocks gorethink.Term) interface{} {
			return DiffPathIndex.Key(file.Commit.Repo.Name, false, file.Path, clocks)
		})
	}).EqJoin(func(x gorethink.Term) gorethink.Term {
		// no-op
		return x
	}, d.getTerm(diffTable), gorethink.EqJoinOpts{
		Index:   DiffPathIndex.GetName(),
		Ordered: true,
	}).Field("right").Run(d.dbClient)
	if err != nil {
		return nil, err
	}

	return d.newFileReader(cursor, file.Path, offset, size), nil
}

type fileReader struct {
	blockClient pfs.BlockAPIClient
	diffs       *gorethink.Cursor
	reader      io.Reader
	offset      int64
	size        int64 // how much data to read
	sizeRead    int64 // how much data has been read
	blockRefs   []*persist.BlockRef
}

func (d *driver) newFileReader(diffs *gorethink.Cursor, path string, offset int64, size int64) *fileReader {
	return &fileReader{
		blockClient: d.blockClient,
		diffs:       diffs,
		offset:      offset,
		size:        size,
	}
}

func (r *fileReader) Read(data []byte) (int, error) {
	var err error
	if r.reader == nil {
		var blockRef *persist.BlockRef
		for {
			for len(r.blockRefs) == 0 {
				// fetch more
				diff := persist.Diff{}
				if r.diffs.Next(&diff) {
					fmt.Printf("diff: %v\n", diff)
					r.blockRefs = append(r.blockRefs, diff.BlockRefs...)
				} else {
					fmt.Println("exiting!")
					if err := r.diffs.Err(); err != nil {
						return 0, err
					}
					return 0, io.EOF
				}
			}
			blockRef = r.blockRefs[0]
			r.blockRefs = r.blockRefs[1:]
			blockSize := int64(blockRef.Size())
			if r.offset >= blockSize {
				r.offset -= blockSize
				continue
			}
			break
		}
		client := client.APIClient{BlockAPIClient: r.blockClient}
		r.reader, err = client.GetBlock(blockRef.Hash, uint64(r.offset), uint64(r.size))
		if err != nil {
			return 0, err
		}
		r.offset = 0
	}
	size, err := r.reader.Read(data)
	if err != nil && err != io.EOF {
		return size, err
	}
	if err == io.EOF {
		r.reader = nil
	}
	r.sizeRead += int64(size)
	if r.sizeRead == r.size {
		return size, io.EOF
	}
	if r.size > 0 && r.sizeRead > r.size {
		return 0, fmt.Errorf("read more than we need; this is likely a bug")
	}
	return size, nil
}

func (r *fileReader) Close() error {
	return r.diffs.Close()
}

func (d *driver) InspectFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64, unsafe bool, handle string) (*pfs.FileInfo, error) {
	return nil, nil
}

func (d *driver) ListFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64, recurse bool, unsafe bool, handle string) ([]*pfs.FileInfo, error) {
	return nil, nil
}

func (d *driver) DeleteFile(file *pfs.File, shard uint64, unsafe bool, handle string) error {
	commit, err := d.getCommitByAmbiguousID(file.Commit.Repo.Name, file.Commit.ID)
	if err != nil {
		return err
	}

	_, err = d.getTerm(diffTable).Insert(&persist.Diff{
		ID:           getDiffID(commit.ID, file.Path),
		CommitID:     commit.ID,
		Repo:         commit.Repo,
		Path:         file.Path,
		BlockRefs:    nil,
		Delete:       true,
		Size:         0,
		BranchClocks: commit.BranchClocks,
	}, gorethink.InsertOpts{
		Conflict: "replace",
	}).RunWrite(d.dbClient)
	return err
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

func (d *driver) insertMessage(table Table, message proto.Message) error {
	_, err := d.getTerm(table).Insert(message).RunWrite(d.dbClient)
	return err
}

func (d *driver) updateMessage(table Table, message proto.Message) error {
	_, err := d.getTerm(table).Insert(message, gorethink.InsertOpts{Conflict: "update"}).RunWrite(d.dbClient)
	return err
}

func (d *driver) getMessageByPrimaryKey(table Table, key interface{}, message proto.Message) error {
	cursor, err := d.getTerm(table).Get(key).Run(d.dbClient)
	if err != nil {
		return err
	}
	err = cursor.One(message)
	if err == gorethink.ErrEmptyResult {
		return fmt.Errorf("%v not found in table %v", key, table)
	}
	return err
}

func (d *driver) getMessageByIndex(table Table, index Index, key interface{}, message proto.Message) error {
	cursor, err := d.getTerm(table).GetAllByIndex(index.GetName(), key).Run(d.dbClient)
	if err != nil {
		return err
	}
	err = cursor.One(message)
	if err == gorethink.ErrEmptyResult {
		return fmt.Errorf("%v not found in index %v of table %v", key, index, table)
	}
	return err
}

// betweenIndex returns a cursor that will return all documents in between two
// values on an index.
// rightBound specifies whether maxVal is included in the range.  Default is false.
func (d *driver) betweenIndex(table Table, index interface{}, minVal interface{}, maxVal interface{}, reverse bool, opts ...gorethink.BetweenOpts) gorethink.Term {
	if reverse {
		index = gorethink.Desc(index)
	}

	return d.getTerm(table).OrderBy(gorethink.OrderByOpts{
		Index: index,
	}).Between(minVal, maxVal, opts...)
}

func (d *driver) deleteMessageByPrimaryKey(table Table, key interface{}) error {
	_, err := d.getTerm(table).Get(key).Delete().RunWrite(d.dbClient)
	return err
}

// OBSOLETE
// Under the new scheme, the parent ID of a commit is self evident:
// The parent of foo/3 is foo/2, for instance.
func (d *driver) getIDOfParentCommit(repo string, commitID string) (string, error) {
	commit, err := d.getCommitByAmbiguousID(repo, commitID)
	if err != nil {
		return "", err
	}
	onlyBranch := commit.BranchClocks[0]
	numClocks := len(onlyBranch.Clocks)
	clock := onlyBranch.Clocks[numClocks-1]
	if clock.Clock == 0 {
		if numClocks < 2 {
			return "", nil
		}
		clock = onlyBranch.Clocks[numClocks-2]
	} else {
		clock.Clock -= 1
	}

	parentCommit := &persist.Commit{}
	if err := d.getMessageByIndex(commitTable, CommitBranchIndex, CommitBranchIndex.Key(commit.Repo, clock.Branch, clock.Clock), parentCommit); err != nil {
		return "", err
	}
	return parentCommit.ID, nil
}

func (d *driver) getCommitByAmbiguousID(repo string, commitID string) (commit *persist.Commit, err error) {
	alias, err := parseClock(commitID)

	commit = &persist.Commit{}
	if err != nil {
		cursor, err := d.getTerm(commitTable).Get(commitID).Run(d.dbClient)
		if err != nil {
			return nil, err
		}

		if err := cursor.One(commit); err != nil {
			return nil, err
		}
	} else {
		if err := d.getMessageByIndex(commitTable, CommitBranchIndex, CommitBranchIndex.Key(repo, alias.Branch, alias.Clock), commit); err != nil {
			return nil, err
		}
	}
	return commit, nil
}

func (d *driver) updateCommitWithAmbiguousID(repo string, commitID string, values map[string]interface{}) (err error) {
	alias, err := parseClock(commitID)
	if err != nil {
		_, err = d.getTerm(commitTable).Get(commitID).Update(values).RunWrite(d.dbClient)
	} else {
		_, err = d.getTerm(commitTable).GetAllByIndex(CommitBranchIndex.GetName(), CommitBranchIndex.Key(repo, alias.Branch, alias.Clock)).Update(values).RunWrite(d.dbClient)
	}
	return err
}
