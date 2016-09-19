package persist

import (
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
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

// ErrCommitNotFound is specified when the commit is not found
type ErrCommitNotFound struct {
	error
}

// ErrBranchExists is specified if the branch already exists
type ErrBranchExists struct {
	error
}

// ErrCommitFinished is specified if the operation is disallowed on Finished Commits
type ErrCommitFinished struct {
	error
}

const (
	repoTable   Table = "Repos"
	diffTable   Table = "Diffs"
	clockTable  Table = "Clocks"
	commitTable Table = "Commits"

	connectTimeoutSeconds = 5
)

const (
	// ErrConflictFileTypeMsg is used when we see a file that is both a file and directory
	ErrConflictFileTypeMsg = "file type conflict"
)

var (
	tables = []Table{
		repoTable,
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

// NewDriver is used to create a new Driver instance
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

// isDBCreated is used to tell when we are trying to initialize a database,
// whether we are getting an error because the database has already been
// initialized.
func isDBCreated(err error) bool {
	return strings.Contains(err.Error(), "Database") && strings.Contains(err.Error(), "already exists")
}

//InitDB is used to setup the database with the tables and indices that PFS requires
func InitDB(address string, dbName string) error {
	session, err := dbConnect(address)
	if err != nil {
		return err
	}
	defer session.Close()

	return initDB(session, dbName)
}

func initDB(session *gorethink.Session, dbName string) error {
	_, err := gorethink.DBCreate(dbName).RunWrite(session)
	if err != nil && !isDBCreated(err) {
		return err
	}

	// Create tables
	for _, table := range tables {
		tableCreateOpts := tableToTableCreateOpts[table]
		if _, err := gorethink.DB(dbName).TableCreate(table, tableCreateOpts...).RunWrite(session); err != nil {
			return err
		}
	}

	// Create indexes
	for _, someIndex := range Indexes {
		if _, err := gorethink.DB(dbName).Table(someIndex.Table).IndexCreateFunc(someIndex.Name, someIndex.CreateFunction, someIndex.CreateOptions).RunWrite(session); err != nil {
			return err
		}
		if _, err := gorethink.DB(dbName).Table(someIndex.Table).IndexWait(someIndex.Name).RunWrite(session); err != nil {
			return err
		}
	}
	return nil
}

// RemoveDB removes the tables in the database that are relavant to PFS
// It keeps the database around tho, as it might contain other tables that
// others created (e.g. PPS).
func RemoveDB(address string, dbName string) error {
	session, err := dbConnect(address)
	if err != nil {
		return err
	}
	defer session.Close()

	return removeDB(session, dbName)
}

func removeDB(session *gorethink.Session, dbName string) error {
	for _, table := range tables {
		if _, err := gorethink.DB(dbName).TableDrop(table).RunWrite(session); err != nil {
			return err
		}
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

func (d *driver) CreateRepo(repo *pfs.Repo, provenance []*pfs.Repo) error {
	if repo == nil {
		return fmt.Errorf("repo cannot be nil")
	}
	err := validateRepoName(repo.Name)
	if err != nil {
		return err
	}

	// Verify that all provenent repos exist
	var provenantRepoNames []interface{}
	var provenantIDs []string
	for _, repo := range provenance {
		provenantRepoNames = append(provenantRepoNames, repo.Name)
		provenantIDs = append(provenantIDs, repo.Name)
	}

	var numProvenantRepos int
	cursor, err := d.getTerm(repoTable).GetAll(provenantRepoNames...).Count().Run(d.dbClient)
	if err != nil {
		return err
	}
	if err := cursor.One(&numProvenantRepos); err != nil {
		return err
	}
	if numProvenantRepos != len(provenance) {
		return fmt.Errorf("could not create repo %v, not all provenance repos exist", repo.Name)
	}

	_, err = d.getTerm(repoTable).Insert(&persist.Repo{
		Name:       repo.Name,
		Created:    now(),
		Provenance: provenantIDs,
	}).RunWrite(d.dbClient)
	if err != nil && gorethink.IsConflictErr(err) {
		return fmt.Errorf("repo %v exists", repo.Name)
	}
	return err
}

func (d *driver) inspectRepo(repo *pfs.Repo) (r *persist.Repo, retErr error) {
	defer func() {
		if retErr == gorethink.ErrEmptyResult {
			retErr = pfsserver.NewErrRepoNotFound(repo.Name)
		}
	}()
	cursor, err := d.getTerm(repoTable).Get(repo.Name).Run(d.dbClient)
	if err != nil {
		return nil, err
	}
	rawRepo := &persist.Repo{}
	if err := cursor.One(rawRepo); err != nil {
		return nil, err
	}
	return rawRepo, nil
}

// byName sorts repos by name
type byName []*pfs.Repo

func (n byName) Len() int           { return len(n) }
func (n byName) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n byName) Less(i, j int) bool { return n[i].Name < n[j].Name }

func (d *driver) InspectRepo(repo *pfs.Repo) (*pfs.RepoInfo, error) {
	rawRepo, err := d.inspectRepo(repo)
	if err != nil {
		return nil, err
	}

	// The full set of this repo's provenance, i.e. its immediate provenance
	// plus the provenance of those repos.
	fullProvenance := make(map[string]bool)
	for _, repoName := range rawRepo.Provenance {
		fullProvenance[repoName] = true
	}

	for _, repoName := range rawRepo.Provenance {
		repoInfo, err := d.InspectRepo(&pfs.Repo{repoName})
		if err != nil {
			return nil, err
		}
		for _, repo := range repoInfo.Provenance {
			fullProvenance[repo.Name] = true
		}
	}

	var provenance []*pfs.Repo
	for repoName := range fullProvenance {
		provenance = append(provenance, &pfs.Repo{
			Name: repoName,
		})
	}

	sort.Sort(byName(provenance))

	return &pfs.RepoInfo{
		Repo: &pfs.Repo{
			Name: rawRepo.Name,
		},
		Created:    rawRepo.Created,
		SizeBytes:  rawRepo.Size,
		Provenance: provenance,
	}, nil
}

func (d *driver) ListRepo(provenance []*pfs.Repo) (repoInfos []*pfs.RepoInfo, retErr error) {
	cursor, err := d.getTerm(repoTable).OrderBy("Name").Run(d.dbClient)
	if err != nil {
		return nil, err
	}
	var repos []*persist.Repo
	if err := cursor.All(&repos); err != nil {
		return nil, err
	}

nextRepo:
	for _, repo := range repos {
		if len(provenance) != 0 {
			// Filter out the repos that don't have the given provenance
			repoInfo, err := d.InspectRepo(&pfs.Repo{repo.Name})
			if err != nil {
				return nil, err
			}
			for _, p := range provenance {
				var found bool
				for _, r := range repoInfo.Provenance {
					if p.Name == r.Name {
						found = true
						break
					}
				}
				if !found {
					continue nextRepo
				}
			}
		}
		repoInfos = append(repoInfos, &pfs.RepoInfo{
			Repo: &pfs.Repo{
				Name: repo.Name,
			},
			Created:   repo.Created,
			SizeBytes: repo.Size,
		})
	}

	return repoInfos, nil
}

func (d *driver) DeleteRepo(repo *pfs.Repo, force bool) error {
	if !force {
		// Make sure that this repo is not the provenance of any other repo
		repoInfos, err := d.ListRepo([]*pfs.Repo{repo})
		if err != nil {
			return err
		}
		if len(repoInfos) > 0 {
			var repoNames []string
			for _, repoInfo := range repoInfos {
				repoNames = append(repoNames, repoInfo.Repo.Name)
			}
			return fmt.Errorf("cannot delete repo %v; it's the provenance of the following repos: %v", repo.Name, repoNames)
		}
	}

	// Deleting in the order of repo -> commits -> diffs seems to make sense,
	// since if one can't see the repo, one can't see the commits; if they can't
	// see the commits, they can't see the diffs.  So in a way we are hiding
	// potential inconsistency here.
	_, err := d.getTerm(repoTable).Get(repo.Name).Delete().RunWrite(d.dbClient)
	if err != nil {
		return err
	}

	_, err = d.getTerm(commitTable).Filter(map[string]interface{}{
		"Repo": repo.Name,
	}).Delete().RunWrite(d.dbClient)
	if err != nil {
		return err
	}

	_, err = d.getTerm(diffTable).Filter(map[string]interface{}{
		"Repo": repo.Name,
	}).Delete().RunWrite(d.dbClient)
	return err
}

func (d *driver) StartCommit(repo *pfs.Repo, commitID string, parentID string, branch string, started *google_protobuf.Timestamp, provenance []*pfs.Commit) (_commit *pfs.Commit, retErr error) {
	if commitID == "" {
		commitID = uuid.NewWithoutDashes()
	}

	rawRepo, err := d.inspectRepo(repo)
	if err != nil {
		return nil, err
	}

	repoSet := make(map[string]bool)
	for _, repoName := range rawRepo.Provenance {
		repoSet[repoName] = true
	}

	var _provenance []*persist.ProvenanceCommit
	for _, c := range provenance {
		if !repoSet[c.Repo.Name] {
			return nil, fmt.Errorf("cannot use %s/%s as provenance, %s is not provenance of %s",
				c.Repo.Name, c.ID, c.Repo.Name, repo.Name)
		}
		_provenance = append(_provenance, &persist.ProvenanceCommit{
			ID:   c.ID,
			Repo: c.Repo.Name,
		})
	}
	// If any of the commit's provenance is archived, the commit should be archived
	var archived bool
	// We compute the complete set of provenance.  That is, the provenance of this
	// commit includes the provenance of its immediate provenance.
	// This is so that running ListCommit with provenance is fast.
	provenanceSet := make(map[string]*pfs.Commit)
	for _, c := range provenance {
		commitInfo, err := d.InspectCommit(c)
		archived = archived || commitInfo.Archived
		if err != nil {
			return nil, err
		}

		for _, p := range commitInfo.Provenance {
			provenanceSet[p.ID] = p
		}
	}

	for _, c := range provenanceSet {
		_provenance = append(_provenance, &persist.ProvenanceCommit{
			ID:   c.ID,
			Repo: c.Repo.Name,
		})
	}

	commit := &persist.Commit{
		ID:         commitID,
		Repo:       repo.Name,
		Started:    now(),
		Provenance: _provenance,
		Archived:   archived,
	}
	var clockID *persist.ClockID
	if parentID == "" {
		if branch == "" {
			branch = uuid.NewWithoutDashes()
		}
		for {
			// The head of this branch will be our parent commit
			parentCommit := &persist.Commit{}
			err := d.getHeadOfBranch(repo.Name, branch, parentCommit)
			if err != nil && err != gorethink.ErrEmptyResult {
				return nil, err
			} else if err == gorethink.ErrEmptyResult {
				// we don't have a parent :(
				// so we create a new clock
				commit.FullClock = append(commit.FullClock, persist.NewClock(branch))
			} else {
				// we do have a parent :D
				// so we inherit our parent's full clock
				// and increment the last component by 1
				commit.FullClock = persist.NewChild(parentCommit.FullClock)
				if err != nil {
					return nil, err
				}
			}
			clock := persist.FullClockHead(commit.FullClock)
			clockID = getClockID(repo.Name, clock)
			err = d.insertMessage(clockTable, clockID)
			if gorethink.IsConflictErr(err) {
				// There is another process creating a commit on this branch
				// at the same time.  We lost the race, but we can try again
				continue
			} else if err != nil {
				return nil, err
			}
			break
		}
	} else {
		parentCommit, err := d.getCommitByAmbiguousID(repo.Name, parentID)
		if err != nil {
			return nil, err
		}

		parentBranch := persist.FullClockBranch(parentCommit.FullClock)

		var newBranch bool
		if branch == "" {
			// Create a commit on the same branch as this parent
			commit.FullClock = persist.NewChild(parentCommit.FullClock)
		} else {
			// Create a new branch based off this parent
			newBranch = true
			commit.FullClock = append(parentCommit.FullClock, persist.NewClock(branch))
			if err != nil {
				return nil, err
			}
		}

		head := persist.FullClockHead(commit.FullClock)
		clockID = getClockID(repo.Name, head)
		if err := d.insertMessage(clockTable, clockID); err != nil {
			if gorethink.IsConflictErr(err) {
				if newBranch {
					// This should only happen if there's another process creating the
					// very same branch at the same time, and we lost the race.
					return nil, ErrBranchExists{fmt.Errorf("branch %s already exists", branch)}
				}
				// This should only happen if there's another process creating a
				// new commit off the same parent, but on the parent's own branch,
				// and we lost the race.
				return nil, fmt.Errorf("%s already has a child on its own branch (%s)", parentID, parentBranch)
			}
			return nil, err
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
	if err := d.insertMessage(commitTable, commit); err != nil {
		return nil, err
	}
	return &pfs.Commit{
		Repo: repo,
		ID:   commitID,
	}, nil
}

func (d *driver) getKey(i *index, desiredOutputDocument interface{}) ([]interface{}, error) {
	cursor, err := gorethink.Expr(i.CreateFunction(gorethink.Expr(desiredOutputDocument))).Run(d.dbClient)
	if err != nil {
		return nil, err
	}
	var key []interface{}
	err = cursor.All(&key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (d *driver) getHeadOfBranch(repo string, branch string, commit *persist.Commit) error {
	startCommit := &persist.Commit{
		Repo: repo,
		FullClock: []*persist.Clock{
			&persist.Clock{
				Branch: branch,
				Clock:  0,
			},
		},
	}
	start, err := d.getKey(CommitClockIndex, startCommit)
	if err != nil {
		return err
	}
	endCommit := &persist.Commit{
		Repo: repo,
		FullClock: []*persist.Clock{
			&persist.Clock{
				Branch: branch,
				Clock:  0,
			},
			&persist.Clock{
				Branch: branch,
				Clock:  math.MaxUint64,
			},
		},
	}
	end, err := d.getKey(CommitClockIndex, endCommit)
	if err != nil {
		return err
	}
	cursor, err := d.betweenIndex(
		commitTable, CommitClockIndex.Name,
		start,
		end,
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
		return nil, fmt.Errorf("invalid commit ID %s", clock)
	}
	c, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid commit ID %s", clock)
	}
	return &persist.Clock{
		Branch: parts[0],
		Clock:  uint64(c),
	}, nil
}

type commitChangeFeed struct {
	NewVal *persist.Commit `gorethink:"new_val,omitempty"`
}

// Given a commitID (database primary key), compute the size of the commit
// using diffs.
func (d *driver) computeCommitSize(commit *persist.Commit) (uint64, error) {
	head := persist.FullClockHead(commit.FullClock)
	cursor, err := d.getTerm(diffTable).GetAllByIndex(
		DiffClockIndex.Name,
		diffClockIndexKey(commit.Repo, head.Branch, head.Clock),
	).Reduce(func(left, right gorethink.Term) gorethink.Term {
		return left.Merge(map[string]interface{}{
			"Size": left.Field("Size").Add(right.Field("Size")),
		})
	}).Default(&persist.Diff{}).Run(d.dbClient)
	if err != nil {
		return 0, err
	}

	var diff persist.Diff
	if err := cursor.One(&diff); err != nil {
		return 0, err
	}

	return diff.Size, nil
}

// FinishCommit blocks until its parent has been finished/cancelled
func (d *driver) FinishCommit(commit *pfs.Commit, finished *google_protobuf.Timestamp, cancel bool) error {
	// TODO: may want to optimize this. Not ideal to jump to DB to validate repo exists. This is required by error strings test in server_test.go
	_, err := d.inspectRepo(commit.Repo)
	if err != nil {
		return err
	}
	rawCommit, err := d.getCommitByAmbiguousID(commit.Repo.Name, commit.ID)
	if err != nil {
		return err
	}

	rawCommit.Size, err = d.computeCommitSize(rawCommit)
	if err != nil {
		return err
	}

	parentID, err := d.getIDOfParentCommit(commit.Repo.Name, commit.ID)
	if err != nil {
		return err
	}

	var parentCancelled bool
	if parentID != "" {
		cursor, err := d.getTerm(commitTable).Get(parentID).Changes(gorethink.ChangesOpts{
			IncludeInitial: true,
		}).Run(d.dbClient)

		if err != nil {
			return err
		}

		var change commitChangeFeed
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

	// Update the size of the repo.  Note that there is a consistency issue here:
	// If this transaction succeeds but the next one (updating Commit) fails,
	// then the repo size will be wrong.  TODO
	_, err = d.getTerm(repoTable).Get(rawCommit.Repo).Update(map[string]interface{}{
		"Size": gorethink.Row.Field("Size").Add(rawCommit.Size),
	}).RunWrite(d.dbClient)
	if err != nil {
		return err
	}

	if finished == nil {
		finished = now()
	}
	rawCommit.Finished = finished
	rawCommit.Cancelled = parentCancelled || cancel
	_, err = d.getTerm(commitTable).Get(rawCommit.ID).Update(rawCommit).RunWrite(d.dbClient)

	return err
}

// ArchiveCommit archives the given commits and all commits that have any of the
// given commits as provenance
func (d *driver) ArchiveCommit(commits []*pfs.Commit) error {
	var commitIDs []interface{}
	for _, commit := range commits {
		c, err := d.getCommitByAmbiguousID(commit.Repo.Name, commit.ID)
		if err != nil {
			return err
		}
		commitIDs = append(commitIDs, c.ID)
	}

	commitIDsTerm := gorethink.Expr(commitIDs)
	query := d.getTerm(commitTable).Filter(func(commit gorethink.Term) gorethink.Term {
		// We want to select all commits that have any of the given commits as
		// provenance
		return gorethink.Or(commit.Field("Provenance").Field("ID").SetIntersection(commitIDsTerm).Count().Ne(0), commitIDsTerm.Contains(commit.Field("ID")))
	}).Update(map[string]interface{}{
		"Archived": true,
	})

	_, err := query.RunWrite(d.dbClient)
	if err != nil {
		return err
	}

	d.getTerm(commitTable).GetAll(commitIDs...)

	return nil
}

func (d *driver) InspectCommit(commit *pfs.Commit) (*pfs.CommitInfo, error) {
	rawCommit, err := d.getCommitByAmbiguousID(commit.Repo.Name, commit.ID)
	if err != nil {
		return nil, err
	}

	commitInfo := d.rawCommitToCommitInfo(rawCommit)
	if commitInfo.Finished == nil {
		commitInfo.SizeBytes, err = d.computeCommitSize(rawCommit)
		if err != nil {
			return nil, err
		}
	}

	// OBSOLETE
	// TestInspectCommit expects request commit ID to match results commit ID
	commitInfo.Commit.ID = commit.ID
	return commitInfo, nil
}

func (d *driver) rawCommitToCommitInfo(rawCommit *persist.Commit) *pfs.CommitInfo {
	commitType := pfs.CommitType_COMMIT_TYPE_READ
	var branch string
	if len(rawCommit.FullClock) > 0 {
		branch = persist.FullClockBranch(rawCommit.FullClock)
	}
	if rawCommit.Finished == nil {
		commitType = pfs.CommitType_COMMIT_TYPE_WRITE
	}

	var provenance []*pfs.Commit
	for _, c := range rawCommit.Provenance {
		provenance = append(provenance, &pfs.Commit{
			Repo: &pfs.Repo{
				Name: c.Repo,
			},
			ID: c.ID,
		})
	}

	// OBSOLETE
	//
	// Here we retrieve the parent commit from the database.
	// This is a HUGE performance issue because we are doing a DB round trip
	// per commit.
	//
	// We do this because some code needs the ParentCommit field of
	// CommitInfo, and they need the ParentCommit to have the actual commit ID.
	//
	// In the future, the client code should be able to directly infer
	// the commit ID (alias) of the parent, e.g. master/1 -> master/0
	parentClock := persist.FullClockParent(rawCommit.FullClock)
	var parentCommit *pfs.Commit
	if parentClock != nil {
		parentClockID := persist.FullClockHead(parentClock).ToCommitID()
		rawParentCommit, _ := d.getCommitByAmbiguousID(rawCommit.Repo, parentClockID)
		parentCommit = &pfs.Commit{
			Repo: &pfs.Repo{
				Name: rawCommit.Repo,
			},
			ID: rawParentCommit.ID,
		}
	}

	return &pfs.CommitInfo{
		Commit: &pfs.Commit{
			Repo: &pfs.Repo{
				Name: rawCommit.Repo,
			},
			ID: rawCommit.ID,
		},
		Branch:       branch,
		Started:      rawCommit.Started,
		Finished:     rawCommit.Finished,
		Cancelled:    rawCommit.Cancelled,
		Archived:     rawCommit.Archived,
		CommitType:   commitType,
		SizeBytes:    rawCommit.Size,
		ParentCommit: parentCommit,
		Provenance:   provenance,
	}
}

func (d *driver) ListCommit(repos []*pfs.Repo, commitType pfs.CommitType, fromCommits []*pfs.Commit, provenance []*pfs.Commit, status pfs.CommitStatus, block bool) ([]*pfs.CommitInfo, error) {
	repoToFromCommit := make(map[string]string)
	for _, repo := range repos {
		// make sure that the repos exist
		_, err := d.inspectRepo(repo)
		if err != nil {
			return nil, err
		}
		repoToFromCommit[repo.Name] = ""
	}
	for _, commit := range fromCommits {
		repoToFromCommit[commit.Repo.Name] = commit.ID
	}
	var queries []interface{}
	for repo, commit := range repoToFromCommit {
		if commit == "" {
			queries = append(queries, d.getTerm(commitTable).OrderBy(gorethink.OrderByOpts{
				Index: CommitFullClockIndex.Name,
			}).Filter(map[string]interface{}{
				"Repo": repo,
			}))
		} else {
			fullClock, err := d.getFullClockByAmbiguousID(repo, commit)
			if err != nil {
				return nil, err
			}
			queries = append(queries, d.getTerm(commitTable).OrderBy(gorethink.OrderByOpts{
				Index: CommitFullClockIndex.Name,
			}).Filter(func(r gorethink.Term) gorethink.Term {
				return gorethink.And(
					r.Field("Repo").Eq(repo),
					r.Field("FullClock").Gt(fullClock),
					// TODO: this is buggy.  "master/0" would be greater than
					// "foo/1" even though "master/0" is not "foo/1"'s descendent.
				)
			}))
		}
	}
	query := gorethink.Union(queries...)
	if status != pfs.CommitStatus_ALL && status != pfs.CommitStatus_CANCELLED {
		query = query.Filter(map[string]interface{}{
			"Cancelled": false,
		})
	}
	if status != pfs.CommitStatus_ALL && status != pfs.CommitStatus_ARCHIVED {
		query = query.Filter(map[string]interface{}{
			"Archived": false,
		})
	}
	switch commitType {
	case pfs.CommitType_COMMIT_TYPE_READ:
		query = query.Filter(func(commit gorethink.Term) gorethink.Term {
			return commit.Field("Finished").Ne(nil)
		})
	case pfs.CommitType_COMMIT_TYPE_WRITE:
		query = query.Filter(func(commit gorethink.Term) gorethink.Term {
			return commit.Field("Finished").Eq(nil)
		})
	}
	var provenanceIDs []interface{}
	for _, commit := range provenance {
		c, err := d.getCommitByAmbiguousID(commit.Repo.Name, commit.ID)
		if err != nil {
			return nil, err
		}
		provenanceIDs = append(provenanceIDs, &persist.ProvenanceCommit{
			ID:   c.ID,
			Repo: c.Repo,
		})
	}
	if provenanceIDs != nil {
		query = query.Filter(func(commit gorethink.Term) gorethink.Term {
			return commit.Field("Provenance").Contains(provenanceIDs...)
		})
	}

	cursor, err := query.Run(d.dbClient)
	if err != nil {
		return nil, err
	}
	var commits []*persist.Commit
	if err := cursor.All(&commits); err != nil {
		return nil, err
	}

	var commitInfos []*pfs.CommitInfo
	if len(commits) > 0 {
		for _, commit := range commits {
			commitInfos = append(commitInfos, d.rawCommitToCommitInfo(commit))
		}
	} else if block {
		query = query.Changes(gorethink.ChangesOpts{
			IncludeInitial: true,
		}).Field("new_val")
		cursor, err := query.Run(d.dbClient)
		if err != nil {
			return nil, err
		}
		var commit persist.Commit
		cursor.Next(&commit)
		if err := cursor.Err(); err != nil {
			return nil, err
		}
		commitInfos = append(commitInfos, d.rawCommitToCommitInfo(&commit))
	}

	return commitInfos, nil
}

func (d *driver) FlushCommit(fromCommits []*pfs.Commit, toRepos []*pfs.Repo) ([]*pfs.CommitInfo, error) {
	repoSet1 := make(map[string]bool)
	for _, commit := range fromCommits {
		repoInfos, err := d.ListRepo([]*pfs.Repo{commit.Repo})
		if err != nil {
			return nil, err
		}
		for _, repoInfo := range repoInfos {
			repoSet1[repoInfo.Repo.Name] = true
		}
	}

	repoSet2 := make(map[string]bool)
	for _, repo := range toRepos {
		repoInfo, err := d.InspectRepo(repo)
		if err != nil {
			return nil, err
		}
		for _, repo := range repoInfo.Provenance {
			repoSet2[repo.Name] = true
		}
		repoSet2[repo.Name] = true
	}

	// The list of the repos that we care about.
	var repos []string
	for repoName := range repoSet1 {
		if len(repoSet2) == 0 || repoSet2[repoName] {
			repos = append(repos, repoName)
		}
	}

	// The commit IDs of the provenance commits
	var provenanceIDs []interface{}
	for _, commit := range fromCommits {
		commit, err := d.getCommitByAmbiguousID(commit.Repo.Name, commit.ID)
		if err != nil {
			return nil, err
		}
		provenanceIDs = append(provenanceIDs, &persist.ProvenanceCommit{
			Repo: commit.Repo,
			ID:   commit.ID,
		})
	}

	if len(provenanceIDs) == 0 {
		return nil, nil
	}

	query := d.getTerm(commitTable).Filter(func(commit gorethink.Term) gorethink.Term {
		return gorethink.And(
			commit.Field("Archived").Eq(false),
			commit.Field("Finished").Ne(nil),
			commit.Field("Provenance").SetIntersection(provenanceIDs).Count().Ne(0),
			gorethink.Expr(repos).Contains(commit.Field("Repo")),
		)
	}).Changes(gorethink.ChangesOpts{
		IncludeInitial: true,
	}).Field("new_val")

	cursor, err := query.Run(d.dbClient)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var commitInfos []*pfs.CommitInfo
	repoSet := make(map[string]bool)
	for _, repoName := range repos {
		repoSet[repoName] = true
	}
	for {
		commit := &persist.Commit{}
		cursor.Next(commit)
		if err := cursor.Err(); err != nil {
			return nil, err
		}
		if commit.Cancelled {
			return commitInfos, fmt.Errorf("commit %s/%s was cancelled", commit.Repo, commit.ID)
		}
		commitInfos = append(commitInfos, d.rawCommitToCommitInfo(commit))
		delete(repoSet, commit.Repo)
		// Return when we have seen at least one commit from each repo that we
		// care about.
		if len(repoSet) == 0 {
			return commitInfos, nil
		}
	}
	panic("unreachable")
}

func (d *driver) ListBranch(repo *pfs.Repo) ([]*pfs.CommitInfo, error) {
	// Get all branches
	cursor, err := d.getTerm(clockTable).OrderBy(gorethink.OrderByOpts{
		Index: ClockBranchIndex.Name,
	}).Between(
		[]interface{}{repo.Name, gorethink.MinVal},
		[]interface{}{repo.Name, gorethink.MaxVal},
	).Field("Branch").Distinct().Run(d.dbClient)
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

func (d *driver) DeleteCommit(commit *pfs.Commit) error {
	return errors.New("DeleteCommit is not implemented")
}

// checkFileType returns an error if the given type conflicts with the preexisting
// type.  TODO: cache file types
func (d *driver) checkFileType(repo string, commit string, path string, typ persist.FileType) (err error) {
	diff, err := d.inspectFile(&pfs.File{
		Commit: &pfs.Commit{
			Repo: &pfs.Repo{
				Name: repo,
			},
			ID: commit,
		},
		Path: path,
	}, nil, nil)
	if err != nil {
		_, ok := err.(*pfsserver.ErrFileNotFound)
		if ok {
			// If the file was not found, then there's no type conflict
			return nil
		}
		return err
	}
	if diff.FileType != typ && diff.FileType != persist.FileType_NONE {
		return errors.New(ErrConflictFileTypeMsg)
	}
	return nil
}

func (d *driver) PutFile(file *pfs.File, delimiter pfs.Delimiter, reader io.Reader) (retErr error) {
	fixPath(file)
	// TODO: eventually optimize this with a cache so that we don't have to
	// go to the database to figure out if the commit exists
	commit, err := d.getCommitByAmbiguousID(file.Commit.Repo.Name, file.Commit.ID)
	if err != nil {
		return err
	}
	if commit.Finished != nil {
		return ErrCommitFinished{fmt.Errorf("commit %v has already been finished", commit.ID)}
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

	var diffs []*persist.Diff
	// the ancestor directories
	for _, prefix := range getPrefixes(file.Path) {
		diffs = append(diffs, &persist.Diff{
			ID:       getDiffID(commit.ID, prefix),
			Repo:     commit.Repo,
			Delete:   false,
			Path:     prefix,
			Clock:    persist.FullClockHead(commit.FullClock),
			FileType: persist.FileType_DIR,
			Modified: now(),
		})
	}

	// the file itself
	diffs = append(diffs, &persist.Diff{
		ID:        getDiffID(commit.ID, file.Path),
		Repo:      commit.Repo,
		Delete:    false,
		Path:      file.Path,
		BlockRefs: refs,
		Size:      size,
		Clock:     persist.FullClockHead(commit.FullClock),
		FileType:  persist.FileType_FILE,
		Modified:  now(),
	})

	// Make sure that there's no type conflict
	for _, diff := range diffs {
		if err := d.checkFileType(commit.Repo, commit.ID, diff.Path, diff.FileType); err != nil {
			return err
		}
	}

	// Actually, we don't know if Rethink actually inserts these documents in
	// order.  If it doesn't, then we might end up with "/foo/bar" but not
	// "/foo", which is kinda problematic.
	_, err = d.getTerm(diffTable).Insert(diffs, gorethink.InsertOpts{
		Conflict: func(id gorethink.Term, oldDoc gorethink.Term, newDoc gorethink.Term) gorethink.Term {
			return gorethink.Branch(
				// We throw an error if the new diff is of a different file type
				// than the old diff, unless the old diff is NONE
				oldDoc.Field("FileType").Ne(persist.FileType_NONE).And(oldDoc.Field("FileType").Ne(newDoc.Field("FileType"))),
				gorethink.Error(ErrConflictFileTypeMsg),
				oldDoc.Merge(map[string]interface{}{
					"BlockRefs": oldDoc.Field("BlockRefs").Add(newDoc.Field("BlockRefs")),
					"Size":      oldDoc.Field("Size").Add(newDoc.Field("Size")),
					// Overwrite the file type in case the old file type is NONE
					"FileType": newDoc.Field("FileType"),
					// Update modification time
					"Modified": newDoc.Field("Modified"),
				}),
			)
		},
	}).RunWrite(d.dbClient)
	return err
}

func now() *google_protobuf.Timestamp {
	return prototime.TimeToTimestamp(time.Now())
}

func getPrefixes(path string) []string {
	prefix := ""
	parts := strings.Split(path, "/")
	var res []string
	// skip the last part; we only want prefixes
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] != "" {
			prefix += "/" + parts[i]
			res = append(res, prefix)
		}
	}
	return res
}

func getDiffID(commitID string, path string) string {
	return fmt.Sprintf("%s:%s", commitID, path)
}

// the equivalent of above except that commitID is a rethink term
func getDiffIDFromTerm(commitID gorethink.Term, path string) gorethink.Term {
	return commitID.Add(":" + path)
}

func (d *driver) MakeDirectory(file *pfs.File) (retErr error) {
	fixPath(file)
	commit, err := d.getCommitByAmbiguousID(file.Commit.Repo.Name, file.Commit.ID)
	if err != nil {
		return err
	}
	if commit.Finished != nil {
		return ErrCommitFinished{fmt.Errorf("commit %v has already been finished", commit.ID)}
	}

	if err := d.checkFileType(commit.Repo, commit.ID, file.Path, persist.FileType_DIR); err != nil {
		return err
	}

	diff := &persist.Diff{
		ID:       getDiffID(commit.ID, file.Path),
		Repo:     commit.Repo,
		Delete:   false,
		Path:     file.Path,
		Clock:    persist.FullClockHead(commit.FullClock),
		FileType: persist.FileType_DIR,
		Modified: now(),
	}
	_, err = d.getTerm(diffTable).Insert(diff).RunWrite(d.dbClient)
	return err
}

func reverseSlice(s []*persist.ClockRange) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// fixPath prepends a slash to the file path if there isn't one,
// and removes the trailing slash if there is one.
func fixPath(file *pfs.File) {
	if len(file.Path) == 0 || file.Path[0] != '/' {
		file.Path = "/" + file.Path
	}
	if len(file.Path) > 1 && file.Path[len(file.Path)-1] == '/' {
		file.Path = file.Path[:len(file.Path)-1]
	}
}

func (d *driver) GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64,
	size int64, diffMethod *pfs.DiffMethod) (io.ReadCloser, error) {
	fixPath(file)
	diff, err := d.inspectFile(file, filterShard, diffMethod)
	if err != nil {
		return nil, err
	}
	if diff.FileType == persist.FileType_DIR {
		return nil, fmt.Errorf("file %s/%s/%s is directory", file.Commit.Repo.Name, file.Commit.ID, file.Path)
	}
	return d.newFileReader(diff.BlockRefs, file, offset, size), nil
}

type fileReader struct {
	blockClient pfs.BlockAPIClient
	reader      io.Reader
	offset      int64
	size        int64 // how much data to read
	sizeRead    int64 // how much data has been read
	blockRefs   []*persist.BlockRef
	file        *pfs.File
}

func (d *driver) newFileReader(blockRefs []*persist.BlockRef, file *pfs.File, offset int64, size int64) *fileReader {
	return &fileReader{
		blockClient: d.blockClient,
		blockRefs:   blockRefs,
		offset:      offset,
		size:        size,
		file:        file,
	}
}

func filterBlockRefs(filterShard *pfs.Shard, file *pfs.File, blockRefs []*persist.BlockRef) []*persist.BlockRef {
	var result []*persist.BlockRef
	for _, blockRef := range blockRefs {
		if pfsserver.BlockInShard(filterShard, file, &pfs.Block{
			Hash: blockRef.Hash,
		}) {
			result = append(result, blockRef)
		}
	}
	return result
}

func (r *fileReader) Read(data []byte) (int, error) {
	var err error
	fmt.Printf("DDD In Read() ... initial r.offset=%v, r.size=%v, r.sizeRead=%v\n", r.offset, r.size, r.sizeRead)
	if r.reader == nil {
		fmt.Printf("DDD Initizializing r.reader\n")
		fmt.Printf("DDD # blockrefs=%v\n", len(r.blockRefs))
		var blockRef *persist.BlockRef
		for {
			fmt.Printf("DDD offset=%v, size=%v\n", r.offset, r.size)
			if len(r.blockRefs) == 0 {
				return 0, io.EOF
			}
			blockRef = r.blockRefs[0]
			fmt.Printf("DDD Assigning current blockref to: %v\n", blockRef)
			r.blockRefs = r.blockRefs[1:]
			blockSize := int64(blockRef.Size())
			fmt.Printf("DDD blocksize=%v\n", blockSize)
			if r.offset >= blockSize {
				r.offset -= blockSize
				continue
			}
			fmt.Printf("DDD offset exceeded blocksize ... breaking\n")
			break
		}
		client := client.APIClient{BlockAPIClient: r.blockClient}
		fmt.Printf("DDD size of reader: %v, read so far: %v\n", r.size, r.sizeRead)
		sizeLeft := r.size
		// e.g. sometimes a reader is constructed of size 0
		if r.size > r.sizeRead {
			//sizeLeft -= r.sizeRead
		}
		fmt.Printf("DDD creating block reader of block %v w offset %v and size %v\n", blockRef.Hash, uint64(r.offset), uint64(r.size-r.sizeRead))
		r.reader, err = client.GetBlock(blockRef.Hash, uint64(r.offset), uint64(sizeLeft))
		//r.reader, err = client.GetBlock(blockRef.Hash, uint64(r.offset), uint64(r.size))
		if err != nil {
			return 0, err
		}
		r.offset = 0
	}
	size, err := r.reader.Read(data)
	fmt.Printf("DDD block reader read %v\n", size)
	if err != nil && err != io.EOF {
		return size, err
	}
	if err == io.EOF {
		if r.sizeRead+int64(size) < r.size {
			// We finished reading this block, but not necessarily all blocks
			// Maybe check if there are remaining blockrefs?
			/*
				fmt.Printf("DDD decrementing size %v by the amount that we read out of the last block %v\n", r.size, r.sizeRead)
				r.size -= r.sizeRead
				fmt.Printf("DDD decremented size: %v\n", r.size)
			*/
		}
		r.reader = nil
	}
	r.sizeRead += int64(size)
	fmt.Printf("DDD incremented sizeRead by size %v to %v\n", size, r.sizeRead)
	if r.sizeRead == r.size {
		return size, io.EOF
	}
	if r.size > 0 && r.sizeRead > r.size {
		fmt.Printf("DDD r.sizeRead=%v, r.size=%v\n", r.sizeRead, r.size)
		return 0, fmt.Errorf("read more than we need; this is likely a bug")
	}
	return size, nil
}

func (r *fileReader) Close() error {
	return nil
}

func (d *driver) InspectFile(file *pfs.File, filterShard *pfs.Shard, diffMethod *pfs.DiffMethod) (*pfs.FileInfo, error) {
	fixPath(file)
	diff, err := d.inspectFile(file, filterShard, diffMethod)
	if err != nil {
		return nil, err
	}

	res := &pfs.FileInfo{
		File: file,
	}

	switch diff.FileType {
	case persist.FileType_FILE:
		res.FileType = pfs.FileType_FILE_TYPE_REGULAR
		res.Modified = diff.Modified

		// OBSOLETE
		// We need the database ID because that's what some old tests expect.
		// Once we switch to semantically meaningful IDs, this won't be necessary.
		commit, err := d.getCommitByAmbiguousID(diff.Repo, diff.CommitID())
		if err != nil {
			return nil, err
		}

		res.CommitModified = &pfs.Commit{
			Repo: file.Commit.Repo,
			ID:   commit.ID,
		}
		res.SizeBytes = diff.Size
	case persist.FileType_DIR:
		res.FileType = pfs.FileType_FILE_TYPE_DIR
		res.Modified = diff.Modified
		childrenDiffs, err := d.getChildren(file.Commit.Repo.Name, file.Path, diffMethod, file.Commit)
		if err != nil {
			return nil, err
		}
		for _, diff := range childrenDiffs {
			res.Children = append(res.Children, &pfs.File{
				Commit: &pfs.Commit{
					Repo: file.Commit.Repo,
					ID:   diff.CommitID(),
				},
				Path: diff.Path,
			})
		}
	case persist.FileType_NONE:
		return nil, pfsserver.NewErrFileNotFound(file.Path, file.Commit.Repo.Name, file.Commit.ID)
	default:
		return nil, fmt.Errorf("unrecognized file type: %d; this is likely a bug", diff.FileType)
	}
	return res, nil
}

func (d *driver) getRangesToMerge(repo string, commits []*pfs.Commit, to string) (*persist.ClockRangeList, error) {
	var ranges persist.ClockRangeList
	for _, commit := range commits {
		clock, err := d.getFullClockByAmbiguousID(commit.Repo.Name, commit.ID)
		if err != nil {
			return nil, err
		}
		ranges.AddFullClock(clock)
	}
	if commit, err := d.getCommitByAmbiguousID(repo, to); err == nil {
		ranges.SubFullClock(commit.FullClock)
	} else if _, ok := err.(*pfsserver.ErrCommitNotFound); !ok {
		return nil, err
	}
	return &ranges, nil
}

// getDiffsToMerge returns the diffs that need to be updated in case of a Merge
func (d *driver) getDiffsToMerge(repo string, commits []*pfs.Commit, to string) (nilTerm gorethink.Term, retErr error) {
	var fullQuery gorethink.Term
	for i, commit := range commits {
		ranges, err := d.getRangesToMerge(repo, []*pfs.Commit{commit}, to)
		if err != nil {
			return nilTerm, err
		}

		var query gorethink.Term
		for j, r := range ranges.Ranges() {
			q := d.getTerm(diffTable).OrderBy(gorethink.OrderByOpts{
				Index: DiffClockIndex.Name,
			}).Between(
				diffClockIndexKey(repo, r.Branch, r.Left),
				diffClockIndexKey(repo, r.Branch, r.Right),
				gorethink.BetweenOpts{
					LeftBound:  "closed",
					RightBound: "closed",
				},
			)
			if j == 0 {
				query = q
			} else {
				query = query.UnionWithOpts(gorethink.UnionOpts{
					Interleave: false,
				}, q)
			}
		}

		query = query.Group("Path").Ungroup().Field("reduction").Map(foldDiffs)
		if i == 0 {
			fullQuery = query
		} else {
			fullQuery = fullQuery.Union(query)
		}
	}

	fullQuery = fullQuery.Group("Path").Ungroup().Field("reduction").Map(foldDiffsWithoutDelete)

	return fullQuery, nil
}

func (d *driver) getCommitsToMerge(repo string, commits []*pfs.Commit, to string) (nilTerm gorethink.Term, retErr error) {
	ranges, err := d.getRangesToMerge(repo, commits, to)
	if err != nil {
		return nilTerm, err
	}

	var query gorethink.Term
	for i, r := range ranges.Ranges() {
		q := d.getTerm(commitTable).OrderBy(gorethink.OrderByOpts{
			Index: CommitClockIndex.Name,
		}).Between(
			commitClockIndexKey(repo, r.Branch, r.Left),
			commitClockIndexKey(repo, r.Branch, r.Right),
			gorethink.BetweenOpts{
				LeftBound:  "closed",
				RightBound: "closed",
			},
		)
		if i == 0 {
			query = q
		} else {
			query = query.UnionWithOpts(gorethink.UnionOpts{
				Interleave: false,
			}, q)
		}
	}
	return query, nil
}

// TODO: rollback
func (d *driver) Merge(repo *pfs.Repo, commits []*pfs.Commit, to string, strategy pfs.MergeStrategy, cancel bool) (retCommits *pfs.Commits, retErr error) {
	// TODO: rollback in the case of a failed merge
	retCommits = &pfs.Commits{
		Commit: []*pfs.Commit{},
	}
	if strategy == pfs.MergeStrategy_SQUASH {
		// Make sure that the commit exists and is open
		commitInfo, err := d.InspectCommit(&pfs.Commit{
			Repo: repo,
			ID:   to,
		})
		if err != nil {
			return nil, err
		}
		if commitInfo.CommitType != pfs.CommitType_COMMIT_TYPE_WRITE {
			return nil, fmt.Errorf("commit %s/%s is not open", repo.Name, to)
		}

		// We first compute the union of the input commits' provenance,
		// which will be the provenance of this merged commit.
		commitsToMerge, err := d.getCommitsToMerge(repo.Name, commits, to)
		if err != nil {
			return nil, err
		}

		cursor, err := commitsToMerge.Map(func(commit gorethink.Term) gorethink.Term {
			return commit.Field("Provenance")
		}).Fold(gorethink.Expr([]interface{}{}), func(acc, provenance gorethink.Term) gorethink.Term {
			return acc.SetUnion(provenance)
		}).Run(d.dbClient)
		if err != nil {
			return nil, err
		}

		var provenanceUnion []*persist.ProvenanceCommit
		if err := cursor.All(&provenanceUnion); err != nil {
			return nil, err
		}

		if _, err := d.getTerm(commitTable).Get(to).Update(map[string]interface{}{
			"Provenance": provenanceUnion,
		}).RunWrite(d.dbClient); err != nil {
			return nil, err
		}

		cursor, err = d.getTerm(commitTable).Get(to).Run(d.dbClient)
		var newPersistCommit persist.Commit
		if err := cursor.One(&newPersistCommit); err != nil {
			return nil, err
		}
		newClock := persist.FullClockHead(newPersistCommit.FullClock)

		diffs, err := d.getDiffsToMerge(repo.Name, commits, to)
		if err != nil {
			return nil, err
		}

		_, err = d.getTerm(diffTable).Insert(diffs.Merge(func(diff gorethink.Term) map[string]interface{} {
			return map[string]interface{}{
				// diff IDs are of the form:
				// commitID:path
				"ID":    gorethink.Expr(to).Add(":", diff.Field("Path")),
				"Clock": newClock,
			}
		})).RunWrite(d.dbClient)
		if err != nil {
			return nil, err
		}

		retCommits.Commit = append(retCommits.Commit, &pfs.Commit{
			Repo: repo,
			ID:   to,
		})
	} else if strategy == pfs.MergeStrategy_REPLAY {
		commits, err := d.getCommitsToMerge(repo.Name, commits, to)
		if err != nil {
			return nil, err
		}

		cursor, err := commits.Run(d.dbClient)
		if err != nil {
			return nil, err
		}

		var rawCommit persist.Commit
		for cursor.Next(&rawCommit) {
			// Copy each commit and their diffs
			// TODO: what if someone else is creating commits on the `to` branch while we
			// are replaying?
			newCommit, err := d.StartCommit(repo, uuid.NewWithoutDashes(), "", to, nil, nil)
			if err != nil {
				return nil, err
			}

			cursor, err := d.getTerm(commitTable).Get(newCommit.ID).Run(d.dbClient)
			var newPersistCommit persist.Commit
			if err := cursor.One(&newPersistCommit); err != nil {
				return nil, err
			}
			newClock := persist.FullClockHead(newPersistCommit.FullClock)
			oldClock := persist.FullClockHead(rawCommit.FullClock)

			// TODO: conflict detection
			_, err = d.getTerm(diffTable).Insert(d.getTerm(diffTable).GetAllByIndex(DiffClockIndex.Name, diffClockIndexKey(repo.Name, oldClock.Branch, oldClock.Clock)).Merge(func(diff gorethink.Term) map[string]interface{} {
				return map[string]interface{}{
					"ID":    gorethink.Expr(newCommit.ID).Add(":", diff.Field("Path")),
					"Clock": newClock,
				}
			})).RunWrite(d.dbClient)
			if err != nil {
				return nil, err
			}

			err = d.FinishCommit(newCommit, nil, cancel)
			retCommits.Commit = append(retCommits.Commit, newCommit)
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unrecognized merge strategy: %v", strategy)
	}

	return retCommits, nil
}

// foldDiffs takes an ordered stream of diffs for a given path, and return
// a single diff that represents the aggregation of these diffs.
func foldDiffs(diffs gorethink.Term) gorethink.Term {
	return diffs.Fold(gorethink.Expr(&persist.Diff{}), func(acc gorethink.Term, diff gorethink.Term) gorethink.Term {
		// TODO: the fold function can easily take offset and size into account,
		// only returning blockrefs that fall into the range specified by offset
		// and size.
		return gorethink.Branch(
			// If neither the acc nor the new diff has FileType_NONE, and they have
			// different FileTypes, then it's a file type conflict.
			acc.Field("FileType").Ne(persist.FileType_NONE).And(diff.Field("FileType").Ne(persist.FileType_NONE).And(acc.Field("FileType").Ne(diff.Field("FileType")))),
			gorethink.Error(ErrConflictFileTypeMsg),
			gorethink.Branch(
				diff.Field("Delete"),
				acc.Merge(diff).Merge(map[string]interface{}{
					"Delete": acc.Field("Delete").Or(diff.Field("Delete")),
				}),
				acc.Merge(diff).Merge(map[string]interface{}{
					"Delete":    acc.Field("Delete").Or(diff.Field("Delete")),
					"BlockRefs": acc.Field("BlockRefs").Add(diff.Field("BlockRefs")),
					"Size":      acc.Field("Size").Add(diff.Field("Size")),
				}),
			),
		)
	})
}

// foldDiffsWithoutDelete is the same as foldDiffs, except that it doesn't remove
// blockrefs ever.
func foldDiffsWithoutDelete(diffs gorethink.Term) gorethink.Term {
	return diffs.Fold(gorethink.Expr(&persist.Diff{}), func(acc gorethink.Term, diff gorethink.Term) gorethink.Term {
		// TODO: the fold function can easily take offset and size into account,
		// only returning blockrefs that fall into the range specified by offset
		// and size.
		return gorethink.Branch(
			// If neither the acc nor the new diff has FileType_NONE, and they have
			// different FileTypes, then it's a file type conflict.
			acc.Field("FileType").Ne(persist.FileType_NONE).And(diff.Field("FileType").Ne(persist.FileType_NONE).And(acc.Field("FileType").Ne(diff.Field("FileType")))),
			gorethink.Error(ErrConflictFileTypeMsg),
			acc.Merge(diff).Merge(map[string]interface{}{
				"Delete":    acc.Field("Delete").Or(diff.Field("Delete")),
				"BlockRefs": acc.Field("BlockRefs").Add(diff.Field("BlockRefs")),
				"Size":      acc.Field("Size").Add(diff.Field("Size")),
			}),
		)
	})
}

func (d *driver) getChildren(repo string, parent string, diffMethod *pfs.DiffMethod, toCommit *pfs.Commit) ([]*persist.Diff, error) {
	query, err := d.getDiffsInCommitRange(diffMethod, toCommit, false, DiffParentIndex.Name, func(clock interface{}) interface{} {
		return diffParentIndexKey(repo, parent, clock)
	})
	if err != nil {
		return nil, err
	}

	cursor, err := query.Group("Path").Ungroup().Field("reduction").Map(foldDiffs).Filter(func(diff gorethink.Term) gorethink.Term {
		return diff.Field("FileType").Ne(persist.FileType_NONE)
	}).OrderBy("Path").Run(d.dbClient)
	if err != nil {
		return nil, err
	}

	var diffs []*persist.Diff
	if err := cursor.All(&diffs); err != nil {
		return nil, err
	}
	return diffs, nil
}

func (d *driver) getChildrenRecursive(repo string, parent string, diffMethod *pfs.DiffMethod, toCommit *pfs.Commit) ([]*persist.Diff, error) {
	query, err := d.getDiffsInCommitRange(diffMethod, toCommit, false, DiffPrefixIndex.Name, func(clock interface{}) interface{} {
		return diffPrefixIndexKey(repo, parent, clock)
	})
	if err != nil {
		return nil, err
	}

	cursor, err := query.Group("Path").Ungroup().Field("reduction").Map(foldDiffs).Filter(func(diff gorethink.Term) gorethink.Term {
		return diff.Field("FileType").Ne(persist.FileType_NONE)
	}).Group(func(diff gorethink.Term) gorethink.Term {
		// This query gives us the first component after the parent prefix.
		// For instance, if the path is "/foo/bar/buzz" and parent is "/foo",
		// this query gives us "bar".
		return gorethink.Branch(
			gorethink.Expr(parent).Eq("/"),
			diff.Field("Path").Split("/").Nth(1),
			diff.Field("Path").Split(parent, 1).Nth(1).Split("/").Nth(1),
		)
	}).Reduce(func(left, right gorethink.Term) gorethink.Term {
		// Basically, we add up the sizes and discard the diff with the longer
		// path.  That way, we will be left with the diff with the shortest path,
		// namely the direct child of parent.
		return gorethink.Branch(
			left.Field("Path").Lt(right.Field("Path")),
			left.Merge(map[string]interface{}{
				"Size": left.Field("Size").Add(right.Field("Size")),
			}),
			right.Merge(map[string]interface{}{
				"Size": left.Field("Size").Add(right.Field("Size")),
			}),
		)
	}).Ungroup().Field("reduction").OrderBy("Path").Run(d.dbClient)
	if err != nil {
		return nil, err
	}

	var diffs []*persist.Diff
	if err := cursor.All(&diffs); err != nil {
		return nil, err
	}

	return diffs, nil
}

type clockToIndexKeyFunc func(interface{}) interface{}

func (d *driver) getDiffsInCommitRange(diffMethod *pfs.DiffMethod, toCommit *pfs.Commit, reverse bool, indexName string, keyFunc clockToIndexKeyFunc) (nilTerm gorethink.Term, retErr error) {
	var from *pfs.Commit
	if diffMethod != nil {
		from = diffMethod.FromCommit
	}
	query, err := d._getDiffsInCommitRange(from, toCommit, reverse, indexName, keyFunc)
	if err != nil {
		return nilTerm, err
	}
	if diffMethod != nil && diffMethod.FullFile {
		// If FullFile is set to true, we first figure out if anything
		// has changed since FromCommit.  If so, we set FromCommit to nil
		cursor, err := query.Count().Gt(0).Run(d.dbClient)
		if err != nil {
			return nilTerm, err
		}
		var hasDiff bool
		if err := cursor.One(&hasDiff); err != nil {
			return nilTerm, err
		}

		if hasDiff {
			from = nil
			query, err = d._getDiffsInCommitRange(from, toCommit, reverse, indexName, keyFunc)
			if err != nil {
				return nilTerm, err
			}
		}
	}
	return query, nil
}

// getDiffsInCommitRange takes a [fromClock, toClock] interval and returns
// an ordered stream of diffs in this range that matches a given index.
// If reverse is set to true, the commits will be in reverse order.
func (d *driver) _getDiffsInCommitRange(fromCommit *pfs.Commit, toCommit *pfs.Commit, reverse bool, indexName string, keyFunc clockToIndexKeyFunc) (nilTerm gorethink.Term, retErr error) {
	var err error
	var fromClock persist.FullClock
	if fromCommit != nil {
		fromClock, err = d.getFullClockByAmbiguousID(fromCommit.Repo.Name, fromCommit.ID)
		if err != nil {
			return nilTerm, err
		}
	}

	toClock, err := d.getFullClockByAmbiguousID(toCommit.Repo.Name, toCommit.ID)
	if err != nil {
		return nilTerm, err
	}

	crl := persist.NewClockRangeList(fromClock, toClock)
	ranges := crl.Ranges()
	if reverse {
		reverseSlice(ranges)
		return gorethink.Expr(ranges).ConcatMap(func(r gorethink.Term) gorethink.Term {
			return d.getTerm(diffTable).OrderBy(gorethink.OrderByOpts{
				Index: gorethink.Desc(indexName),
			}).Between(
				keyFunc([]interface{}{r.Field("Branch"), r.Field("Left")}),
				keyFunc([]interface{}{r.Field("Branch"), r.Field("Right")}),
				gorethink.BetweenOpts{
					LeftBound:  "closed",
					RightBound: "closed",
				},
			)
		}), nil
	}
	return gorethink.Expr(ranges).ConcatMap(func(r gorethink.Term) gorethink.Term {
		return d.getTerm(diffTable).OrderBy(gorethink.OrderByOpts{
			Index: indexName,
		}).Between(
			keyFunc([]interface{}{r.Field("Branch"), r.Field("Left")}),
			keyFunc([]interface{}{r.Field("Branch"), r.Field("Right")}),
			gorethink.BetweenOpts{
				LeftBound:  "closed",
				RightBound: "closed",
			},
		)
	}), nil
}

func (d *driver) getFullClockByAmbiguousID(repo string, commitID string) (persist.FullClock, error) {
	commit, err := d.getCommitByAmbiguousID(repo, commitID)
	if err != nil {
		return nil, err
	}
	return commit.FullClock, nil
}

func (d *driver) inspectFile(file *pfs.File, filterShard *pfs.Shard, diffMethod *pfs.DiffMethod) (*persist.Diff, error) {
	if !pfsserver.FileInShard(filterShard, file) {
		return nil, pfsserver.NewErrFileNotFound(file.Path, file.Commit.Repo.Name, file.Commit.ID)
	}

	query, err := d.getDiffsInCommitRange(diffMethod, file.Commit, false, DiffPathIndex.Name, func(clock interface{}) interface{} {
		return diffPathIndexKey(file.Commit.Repo.Name, file.Path, clock)
	})
	if err != nil {
		return nil, err
	}

	cursor, err := foldDiffs(query).Run(d.dbClient)
	if err != nil {
		return nil, err
	}

	diff := &persist.Diff{}
	if err := cursor.One(diff); err != nil {
		if err == gorethink.ErrEmptyResult {
			return nil, pfsserver.NewErrFileNotFound(file.Path, file.Commit.Repo.Name, file.Commit.ID)
		}
		return nil, err
	}

	if len(diff.BlockRefs) == 0 {
		// If the file is empty, we want to make sure that it's seen by one shard.
		if !pfsserver.BlockInShard(filterShard, file, nil) {
			return nil, pfsserver.NewErrFileNotFound(file.Path, file.Commit.Repo.Name, file.Commit.ID)
		}
	} else {
		// If the file is not empty, we want to make sure to return NotFound if
		// all blocks have been filtered out.
		diff.BlockRefs = filterBlockRefs(filterShard, file, diff.BlockRefs)
		if len(diff.BlockRefs) == 0 {
			return nil, pfsserver.NewErrFileNotFound(file.Path, file.Commit.Repo.Name, file.Commit.ID)
		}
	}

	return diff, nil
}

func (d *driver) ListFile(file *pfs.File, filterShard *pfs.Shard, diffMethod *pfs.DiffMethod, recurse bool) ([]*pfs.FileInfo, error) {
	fixPath(file)
	// We treat the root directory specially: we know that it's a directory
	if file.Path != "/" {
		fileInfo, err := d.InspectFile(file, filterShard, diffMethod)
		if err != nil {
			return nil, err
		}
		switch fileInfo.FileType {
		case pfs.FileType_FILE_TYPE_REGULAR:
			return []*pfs.FileInfo{fileInfo}, nil
		case pfs.FileType_FILE_TYPE_DIR:
			break
		default:
			return nil, fmt.Errorf("unrecognized file type %d; this is likely a bug", fileInfo.FileType)
		}
	}

	var diffs []*persist.Diff
	var err error
	if recurse {
		diffs, err = d.getChildrenRecursive(file.Commit.Repo.Name, file.Path, diffMethod, file.Commit)
	} else {
		diffs, err = d.getChildren(file.Commit.Repo.Name, file.Path, diffMethod, file.Commit)
	}
	if err != nil {
		return nil, err
	}

	var fileInfos []*pfs.FileInfo
	for _, diff := range diffs {
		fileInfo := &pfs.FileInfo{}
		fileInfo.File = &pfs.File{
			Commit: file.Commit,
			Path:   diff.Path,
		}
		fileInfo.SizeBytes = diff.Size
		fileInfo.Modified = diff.Modified
		switch diff.FileType {
		case persist.FileType_FILE:
			fileInfo.FileType = pfs.FileType_FILE_TYPE_REGULAR
		case persist.FileType_DIR:
			fileInfo.FileType = pfs.FileType_FILE_TYPE_DIR
		default:
			return nil, fmt.Errorf("unrecognized file type %d; this is likely a bug", diff.FileType)
		}
		fileInfo.CommitModified = &pfs.Commit{
			Repo: file.Commit.Repo,
			ID:   diff.CommitID(),
		}
		// TODO - This filtering should be done at the DB level
		if pfsserver.FileInShard(filterShard, fileInfo.File) {
			fileInfos = append(fileInfos, fileInfo)
		}
	}

	return fileInfos, nil
}

func (d *driver) DeleteFile(file *pfs.File) error {
	fixPath(file)

	commit, err := d.getCommitByAmbiguousID(file.Commit.Repo.Name, file.Commit.ID)
	if err != nil {
		return err
	}

	repo := commit.Repo
	commitID := commit.ID
	prefix := file.Path

	query, err := d.getDiffsInCommitRange(nil, file.Commit, false, DiffPrefixIndex.Name, func(clock interface{}) interface{} {
		return diffPrefixIndexKey(repo, prefix, clock)
	})
	if err != nil {
		return err
	}

	// Get all files under the directory, ordered by path.
	cursor, err := query.Group("Path").Ungroup().Field("reduction").Map(foldDiffs).Filter(func(diff gorethink.Term) gorethink.Term {
		return diff.Field("FileType").Ne(persist.FileType_NONE)
	}).Field("Path").Run(d.dbClient)
	if err != nil {
		return err
	}

	var paths []string
	if err := cursor.All(&paths); err != nil {
		return err
	}
	paths = append(paths, prefix)

	var diffs []*persist.Diff
	for _, path := range paths {
		diffs = append(diffs, &persist.Diff{
			ID:        getDiffID(commitID, path),
			Repo:      repo,
			Path:      path,
			BlockRefs: nil,
			Delete:    true,
			Size:      0,
			Clock:     persist.FullClockHead(commit.FullClock),
			FileType:  persist.FileType_NONE,
		})
	}

	// TODO: ideally we want to insert the documents ordered by their path,
	// where we insert the leaves first all the way to the root.  That way
	// we ensure the consistency of the file system: it's ok if we've removed
	// "/foo/bar" but not "/foo", but it's problematic if we've removed "/foo"
	// but not "/foo/bar"
	_, err = d.getTerm(diffTable).Insert(diffs, gorethink.InsertOpts{
		Conflict: "replace",
	}).RunWrite(d.dbClient)

	return err
}

func (d *driver) DeleteAll() error {
	for _, table := range tables {
		if _, err := d.getTerm(table).Delete().RunWrite(d.dbClient); err != nil {
			return err
		}
	}

	return nil
}

func (d *driver) ArchiveAll() error {
	_, err := d.getTerm(commitTable).Update(map[string]interface{}{
		"Archived": true,
	}).RunWrite(d.dbClient)
	return err
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

func (d *driver) getMessageByIndex(table Table, i *index, key interface{}, message proto.Message) error {
	cursor, err := d.getTerm(table).GetAllByIndex(i.Name, key).Run(d.dbClient)
	if err != nil {
		return err
	}
	err = cursor.One(message)
	if err == gorethink.ErrEmptyResult {
		return fmt.Errorf("%v not found in index %v of table %v", key, i, table)
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

func (d *driver) getIDOfParentCommit(repo string, commitID string) (string, error) {
	commit, err := d.getCommitByAmbiguousID(repo, commitID)
	if err != nil {
		return "", err
	}
	clock := persist.FullClockHead(commit.FullClock)
	if clock.Clock == 0 {
		// e.g. the parent of [(master, 1), (foo, 0)] is [(master, 1)]
		if len(commit.FullClock) < 2 {
			return "", nil
		}
		clock = commit.FullClock[len(commit.FullClock)-2]
	} else {
		clock.Clock--
	}

	parentCommit := &persist.Commit{}
	if err := d.getMessageByIndex(commitTable, CommitClockIndex, commitClockIndexKey(commit.Repo, clock.Branch, clock.Clock), parentCommit); err != nil {
		return "", err
	}
	return parentCommit.ID, nil
}

// getCommitByAmbiguousID accepts a repo name and an ID, and returns a Commit object.
// The ID can be of 3 forms:
// 1. Database primary key: we are only supporting this case to maintain compatibility
// of the existing tests.  We will remove support for this case eventually.  OBSOLETE
// 2. branch/clock: like "master/3"
// 3. branch: like "master".  This would represent the head of the branch.
func (d *driver) getCommitByAmbiguousID(repo string, commitID string) (commit *persist.Commit, retErr error) {
	defer func() {
		if retErr == gorethink.ErrEmptyResult {
			retErr = pfsserver.NewErrCommitNotFound(repo, commitID)
		}
	}()
	alias, err := parseClock(commitID)

	commit = &persist.Commit{}
	if err != nil {
		// We see if the commitID is a branch name
		if err := d.getHeadOfBranch(repo, commitID, commit); err != nil {
			if err != gorethink.ErrEmptyResult {
				return nil, err
			}

			// If the commit ID is not a branch name, we see if it's a database key
			cursor, err := d.getTerm(commitTable).Get(commitID).Run(d.dbClient)
			if err != nil {
				return nil, err
			}

			if err := cursor.One(commit); err != nil {
				return nil, err
			}
		}
	} else {
		// If we can't parse
		if err := d.getMessageByIndex(commitTable, CommitClockIndex, commitClockIndexKey(repo, alias.Branch, alias.Clock), commit); err != nil {
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
		_, err = d.getTerm(commitTable).GetAllByIndex(CommitClockIndex.Name, commitClockIndexKey(repo, alias.Branch, alias.Clock)).Update(values).RunWrite(d.dbClient)
	}
	return err
}
