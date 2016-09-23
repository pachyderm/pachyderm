package persist

import (
	"crypto/sha256"
	"encoding/hex"
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
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"

	"github.com/dancannon/gorethink"
	"github.com/gogo/protobuf/proto"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/time"
	"google.golang.org/grpc"
)

// A Table is a rethinkdb table name.
type Table string

// A PrimaryKey is a rethinkdb primary key identifier.
type PrimaryKey string

const (
	repoTable   Table = "Repos"
	diffTable   Table = "Diffs"
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

func (d *driver) getFullProvenance(repo *pfs.Repo, provenance []*pfs.Commit) (fullProvenance []*persist.ProvenanceCommit, archived bool, err error) {
	rawRepo, err := d.inspectRepo(repo)
	if err != nil {
		return nil, false, err
	}

	repoSet := make(map[string]bool)
	for _, repoName := range rawRepo.Provenance {
		repoSet[repoName] = true
	}

	for _, c := range provenance {
		if !repoSet[c.Repo.Name] {
			return nil, false, fmt.Errorf("cannot use %s/%s as provenance, %s is not provenance of %s", c.Repo.Name, c.ID, c.Repo.Name, repo.Name)
		}
		fullProvenance = append(fullProvenance, &persist.ProvenanceCommit{
			ID:   c.ID,
			Repo: c.Repo.Name,
		})
	}
	// We compute the complete set of provenance.  That is, the provenance of this
	// commit includes the provenance of its immediate provenance.
	// This is so that running ListCommit with provenance is fast.
	provenanceSet := make(map[string]*pfs.Commit)
	for _, c := range provenance {
		commitInfo, err := d.InspectCommit(c)
		// If any of the commit's provenance is archived, the commit should be
		// archived
		archived = archived || commitInfo.Archived
		if err != nil {
			return nil, false, err
		}

		for _, p := range commitInfo.Provenance {
			provenanceSet[p.ID] = p
		}
	}

	for _, c := range provenanceSet {
		fullProvenance = append(fullProvenance, &persist.ProvenanceCommit{
			ID:   c.ID,
			Repo: c.Repo.Name,
		})
	}

	return fullProvenance, archived, nil
}

func (d *driver) ForkCommit(parent *pfs.Commit, branch string, provenance []*pfs.Commit) (*pfs.Commit, error) {
	fullProvenance, archived, err := d.getFullProvenance(parent.Repo, provenance)
	if err != nil {
		return nil, err
	}

	clock := persist.NewClock(branch)
	commit := &persist.Commit{
		ID:         persist.NewCommitID(parent.Repo.Name, clock),
		Repo:       parent.Repo.Name,
		Started:    now(),
		Provenance: fullProvenance,
		Archived:   archived,
	}

	parentCommit, err := d.getRawCommit(parent)
	if err != nil {
		return nil, err
	}

	commit.FullClock = append(parentCommit.FullClock, persist.NewClock(branch))

	if err := d.insertMessage(commitTable, commit); err != nil {
		if gorethink.IsConflictErr(err) {
			return nil, pfsserver.NewErrCommitExists(commit.Repo, commit.ID)
		}
		return nil, err
	}

	return &pfs.Commit{
		Repo: parent.Repo,
		ID:   clock.ReadableCommitID(),
	}, nil
}

func isBranchName(id string) bool {
	return !strings.Contains(id, "/")
}

func (d *driver) StartCommit(parent *pfs.Commit, provenance []*pfs.Commit) (*pfs.Commit, error) {
	if parent.Repo.Name == "" || parent.ID == "" {
		return nil, fmt.Errorf("Invalid parent commit: %s/%s", parent.Repo.Name, parent.ID)
	}

	fullProvenance, archived, err := d.getFullProvenance(parent.Repo, provenance)
	if err != nil {
		return nil, err
	}

	commit := &persist.Commit{
		Repo:       parent.Repo.Name,
		Started:    now(),
		Provenance: fullProvenance,
		Archived:   archived,
	}

	var makeNewBranch bool
	parentCommit, err := d.getRawCommit(parent)
	if _, ok := err.(*pfsserver.ErrCommitNotFound); ok && isBranchName(parent.ID) {
		makeNewBranch = true
	} else if err == gorethink.ErrEmptyResult {
		return nil, pfsserver.NewErrCommitNotFound(parent.Repo.Name, parent.ID)
	} else if err != nil {
		return nil, err
	}

	var clock *persist.Clock
	if makeNewBranch {
		branch := parent.ID
		clock = persist.NewClock(branch)
		commit.ID = persist.NewCommitID(parent.Repo.Name, clock)
		commit.FullClock = append(commit.FullClock, clock)
	} else {
		commit.FullClock = persist.NewChild(parentCommit.FullClock)
		clock = persist.FullClockHead(commit.FullClock)
		commit.ID = persist.NewCommitID(parent.Repo.Name, clock)
	}

	if err := d.insertMessage(commitTable, commit); err != nil {
		// TODO: there can be a race if two threads concurrently start commit
		// using a branch name.  We should automatically detect the race and retry.
		if gorethink.IsConflictErr(err) {
			return nil, pfsserver.NewErrCommitExists(commit.Repo, commit.ID)
		}
		return nil, err
	}

	return &pfs.Commit{
		Repo: parent.Repo,
		ID:   clock.ReadableCommitID(),
	}, nil
}

func getRawCommitID(repo string, readableCommitID string) (string, error) {
	c, err := parseClock(readableCommitID)
	if err != nil {
		return "", err
	}
	return persist.NewCommitID(repo, c), nil
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

func (d *driver) getHeadOfBranch(repo string, branch string, retCommit *persist.Commit) error {
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
	return cursor.One(retCommit)
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
func (d *driver) FinishCommit(commit *pfs.Commit, cancel bool) error {
	// TODO: may want to optimize this. Not ideal to jump to DB to validate repo exists. This is required by error strings test in server_test.go
	_, err := d.inspectRepo(commit.Repo)
	if err != nil {
		return err
	}
	rawCommit, err := d.getRawCommit(commit)
	if err != nil {
		return err
	}

	rawCommit.Size, err = d.computeCommitSize(rawCommit)
	if err != nil {
		return err
	}

	parentClock := persist.FullClockParent(rawCommit.FullClock)
	var parentCancelled bool
	if parentClock != nil {
		parentID := persist.NewCommitID(rawCommit.Repo, persist.FullClockHead(parentClock))
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

	rawCommit.Finished = now()
	rawCommit.Cancelled = parentCancelled || cancel
	_, err = d.getTerm(commitTable).Get(rawCommit.ID).Update(rawCommit).RunWrite(d.dbClient)

	return err
}

// ArchiveCommits archives the given commits and all commits that have any of the
// given commits as provenance
func (d *driver) ArchiveCommit(commits []*pfs.Commit) error {
	var provenanceIDs []interface{}
	for _, commit := range commits {
		provenanceIDs = append(provenanceIDs, commit.ID)
	}

	var rawIDs []interface{}
	for _, commit := range commits {
		rawID, err := getRawCommitID(commit.Repo.Name, commit.ID)
		if err != nil {
			return err
		}
		rawIDs = append(rawIDs, rawID)
	}

	query := d.getTerm(commitTable).Filter(func(commit gorethink.Term) gorethink.Term {
		// We want to select all commits that have any of the given commits as
		// provenance
		return gorethink.Or(commit.Field("Provenance").Field("ID").SetIntersection(gorethink.Expr(provenanceIDs)).Count().Ne(0), gorethink.Expr(rawIDs).Contains(commit.Field("ID")))
	}).Update(map[string]interface{}{
		"Archived": true,
	})

	_, err := query.RunWrite(d.dbClient)
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) InspectCommit(commit *pfs.Commit) (*pfs.CommitInfo, error) {
	rawCommit, err := d.getRawCommit(commit)
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

	parentClock := persist.FullClockParent(rawCommit.FullClock)
	var parentCommit *pfs.Commit
	if parentClock != nil {
		parentCommit = &pfs.Commit{
			Repo: &pfs.Repo{
				Name: rawCommit.Repo,
			},
			ID: persist.FullClockHead(parentClock).ReadableCommitID(),
		}
	}

	return &pfs.CommitInfo{
		Commit: &pfs.Commit{
			Repo: &pfs.Repo{
				Name: rawCommit.Repo,
			},
			ID: persist.FullClockHead(rawCommit.FullClock).ReadableCommitID(),
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

func (d *driver) ListCommit(fromCommits []*pfs.Commit, provenance []*pfs.Commit, commitType pfs.CommitType, status pfs.CommitStatus, block bool) ([]*pfs.CommitInfo, error) {
	repoToFromCommit := make(map[string]string)
	for _, commit := range fromCommits {
		// make sure that the repos exist
		_, err := d.inspectRepo(commit.Repo)
		if err != nil {
			return nil, err
		}
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
			fullClock, err := d.getFullClock(&pfs.Commit{
				Repo: &pfs.Repo{repo},
				ID:   commit,
			})
			if err != nil {
				return nil, err
			}
			queries = append(queries, d.getTerm(commitTable).OrderBy(gorethink.OrderByOpts{
				Index: CommitFullClockIndex.Name,
			}).Filter(func(r gorethink.Term) gorethink.Term {
				return gorethink.And(
					r.Field("Repo").Eq(repo),
					persist.DBClockDescendent(r.Field("FullClock"), gorethink.Expr(fullClock)),
				)
			}))
		}
	}

	var query gorethink.Term
	if len(queries) > 0 {
		query = gorethink.Union(queries...)
	} else {
		query = d.getTerm(commitTable).OrderBy(gorethink.OrderByOpts{
			Index: CommitFullClockIndex.Name,
		})
	}

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
		// TODO: we need to validate the provenanceIDs: 1) they must actually
		// exist, and 2) they can't be just branch names
		provenanceIDs = append(provenanceIDs, &persist.ProvenanceCommit{
			ID:   commit.ID,
			Repo: commit.Repo.Name,
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

	var result []*pfs.CommitInfo
	// The commit IDs of the provenance commits
	var provenanceIDs []interface{}
	for _, commit := range fromCommits {
		rawCommit, err := d.getRawCommit(commit)
		if err != nil {
			return nil, err
		}
		result = append(result, d.rawCommitToCommitInfo(rawCommit))
		provenanceIDs = append(provenanceIDs, &persist.ProvenanceCommit{
			Repo: commit.Repo.Name,
			// We can't just use commit.ID directly because it might be a
			// branch name instead of a commitID
			ID: persist.FullClockHead(rawCommit.FullClock).ReadableCommitID(),
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

	repoSet := make(map[string]bool)
	for _, repoName := range repos {
		repoSet[repoName] = true
	}
	for len(repoSet) > 0 {
		commit := &persist.Commit{}
		cursor.Next(commit)
		if err := cursor.Err(); err != nil {
			return nil, err
		}
		if commit.Cancelled {
			return nil, fmt.Errorf("commit %s/%s was cancelled", commit.Repo, commit.ID)
		}
		result = append(result, d.rawCommitToCommitInfo(commit))
		delete(repoSet, commit.Repo)
	}
	return result, nil
}

func (d *driver) ListBranch(repo *pfs.Repo, status pfs.CommitStatus) ([]string, error) {
	// Get all branches
	cursor, err := d.getTerm(commitTable).Distinct(gorethink.DistinctOpts{
		Index: CommitBranchIndex.Name,
	}).Filter(gorethink.Row.Nth(0).Eq(repo.Name)).OrderBy(gorethink.Row.Nth(1)).Map(gorethink.Row.Nth(1)).Run(d.dbClient)
	if err != nil {
		return nil, err
	}

	var branches []string
	if err := cursor.All(&branches); err != nil {
		return nil, err
	}

	if status == pfs.CommitStatus_ALL {
		return branches, nil
	}

	var res []string
	for _, branch := range branches {
		commit := &persist.Commit{}
		if err := d.getHeadOfBranch(repo.Name, branch, commit); err != nil {
			return nil, err
		}
		// Check if we should skip the commit based on status
		if status != pfs.CommitStatus_ALL {
			if commit.Cancelled && status != pfs.CommitStatus_CANCELLED {
				continue
			}
			if commit.Archived && status != pfs.CommitStatus_ARCHIVED {
				continue
			}
		}
		res = append(res, branch)
	}
	return res, nil
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

// checkPath checks if a file path is legal
func checkPath(path string) error {
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("filename cannot contain null character: %s", path)
	}
	return nil
}

func (d *driver) PutFile(file *pfs.File, delimiter pfs.Delimiter, reader io.Reader) (retErr error) {
	fixPath(file)
	if err := checkPath(file.Path); err != nil {
		return err
	}

	// TODO: eventually optimize this with a cache so that we don't have to
	// go to the database to figure out if the commit exists
	commit, err := d.getRawCommit(file.Commit)
	if err != nil {
		return err
	}
	if commit.Finished != nil {
		return pfsserver.NewErrCommitFinished(commit.Repo, commit.ID)
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
			ID:       getDiffID(commit.Repo, commit.ID, prefix),
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
		ID:        getDiffID(commit.Repo, commit.ID, file.Path),
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
		if err := d.checkFileType(file.Commit.Repo.Name, file.Commit.ID, diff.Path, diff.FileType); err != nil {
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

func getDiffID(repo string, commitID string, path string) string {
	s := fmt.Sprintf("%s:%s:%s", repo, commitID, path)
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

func (d *driver) MakeDirectory(file *pfs.File) (retErr error) {
	fixPath(file)
	commit, err := d.getRawCommit(file.Commit)
	if err != nil {
		return err
	}
	if commit.Finished != nil {
		return pfsserver.NewErrCommitFinished(commit.Repo, commit.ID)
	}

	if err := d.checkFileType(file.Commit.Repo.Name, file.Commit.ID, file.Path, persist.FileType_DIR); err != nil {
		return err
	}

	diff := &persist.Diff{
		ID:       getDiffID(commit.Repo, commit.ID, file.Path),
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

// filterBlocks filters out blockrefs for a given diff, or return a FileNotFound
// error if all of the blockrefs have been figured out, except that we want to
// make sure that there's at least one shard that matches a given empty diff
func filterBlocks(diff *persist.Diff, filterShard *pfs.Shard, file *pfs.File) (*persist.Diff, error) {
	if len(diff.BlockRefs) == 0 {
		// If the file is empty, we want to make sure that it's seen by one shard.
		if !pfsserver.BlockInShard(filterShard, file, nil) {
			return nil, pfsserver.NewErrFileNotFound(file.Path, file.Commit.Repo.Name, file.Commit.ID)
		}
	} else {
		// If the file is not empty, we want to make sure to return NotFound if
		// all blocks have been filtered out.
		var result []*persist.BlockRef
		for _, blockRef := range diff.BlockRefs {
			if pfsserver.BlockInShard(filterShard, file, &pfs.Block{
				Hash: blockRef.Hash,
			}) {
				result = append(result, blockRef)
			}
		}
		diff.BlockRefs = result
		if len(diff.BlockRefs) == 0 {
			return nil, pfsserver.NewErrFileNotFound(file.Path, file.Commit.Repo.Name, file.Commit.ID)
		}
		var size uint64
		for _, blockref := range diff.BlockRefs {
			size += blockref.Size()
		}
		diff.Size = size
	}
	return diff, nil
}

func (r *fileReader) Read(data []byte) (int, error) {
	var err error
	if r.reader == nil {
		var blockRef *persist.BlockRef
		for {
			if len(r.blockRefs) == 0 {
				return 0, io.EOF
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
		sizeLeft := r.size
		// e.g. sometimes a reader is constructed of size 0
		if sizeLeft != 0 {
			sizeLeft -= r.sizeRead
		}
		r.reader, err = client.GetBlock(blockRef.Hash, uint64(r.offset), uint64(sizeLeft))
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

		res.CommitModified = &pfs.Commit{
			Repo: file.Commit.Repo,
			ID:   diff.Clock.ReadableCommitID(),
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

func (d *driver) getRangesToMerge(commits []*pfs.Commit, to *pfs.Commit) (*persist.ClockRangeList, error) {
	var ranges persist.ClockRangeList
	for _, commit := range commits {
		clock, err := d.getFullClock(commit)
		if err != nil {
			return nil, err
		}
		ranges.AddFullClock(clock)
	}
	if commit, err := d.getRawCommit(to); err == nil {
		ranges.SubFullClock(commit.FullClock)
	} else if _, ok := err.(*pfsserver.ErrCommitNotFound); !ok {
		return nil, err
	}
	return &ranges, nil
}

// getDiffsToMerge returns the diffs that need to be updated in case of a Merge
func (d *driver) getDiffsToMerge(commits []*pfs.Commit, to *pfs.Commit) (nilTerm gorethink.Term, retErr error) {
	repo := to.Repo.Name
	var fullQuery gorethink.Term
	for i, commit := range commits {
		ranges, err := d.getRangesToMerge([]*pfs.Commit{commit}, to)
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

func (d *driver) getCommitsToMerge(commits []*pfs.Commit, to *pfs.Commit) (nilTerm gorethink.Term, retErr error) {
	repo := to.Repo.Name
	ranges, err := d.getRangesToMerge(commits, to)
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

func (d *driver) SquashCommit(fromCommits []*pfs.Commit, toCommit *pfs.Commit) error {
	if len(fromCommits) == 0 || toCommit == nil {
		return fmt.Errorf("Invalid arguments: fromCommits: %v; toCommit: %v", fromCommits, toCommit)
	}

	// Make sure that the commit exists and is open
	commitInfo, err := d.InspectCommit(toCommit)
	if err != nil {
		return err
	}
	if commitInfo.CommitType != pfs.CommitType_COMMIT_TYPE_WRITE {
		return fmt.Errorf("commit %s/%s is not open", toCommit.Repo.Name, toCommit.ID)
	}

	// Make sure that all fromCommits are from the same repo
	var repo string
	for _, commit := range fromCommits {
		if repo == "" {
			repo = commit.Repo.Name
		} else {
			if repo != commit.Repo.Name {
				return fmt.Errorf("you can only squash commits from the same repo; found %s and %s", repo, commit.Repo.Name)
			}
		}
	}

	if repo != toCommit.Repo.Name {
		return fmt.Errorf("you can only squash commits into the same repo; found %s and %s", repo, toCommit.Repo.Name)
	}

	// We first compute the union of the input commits' provenance,
	// which will be the provenance of this merged commit.
	commitsToMerge, err := d.getCommitsToMerge(fromCommits, toCommit)
	if err != nil {
		return err
	}

	cursor, err := commitsToMerge.Map(func(commit gorethink.Term) gorethink.Term {
		return commit.Field("Provenance")
	}).Fold(gorethink.Expr([]interface{}{}), func(acc, provenance gorethink.Term) gorethink.Term {
		return acc.SetUnion(provenance)
	}).Run(d.dbClient)
	if err != nil {
		return err
	}

	var provenanceUnion []*persist.ProvenanceCommit
	if err := cursor.All(&provenanceUnion); err != nil {
		return err
	}

	rawCommitID, err := getRawCommitID(toCommit.Repo.Name, toCommit.ID)
	if err != nil {
		return err
	}

	if _, err := d.getTerm(commitTable).Get(rawCommitID).Update(map[string]interface{}{
		"Provenance": provenanceUnion,
	}).RunWrite(d.dbClient); err != nil {
		return err
	}

	cursor, err = d.getTerm(commitTable).Get(rawCommitID).Run(d.dbClient)
	var newPersistCommit persist.Commit
	if err := cursor.One(&newPersistCommit); err != nil {
		return err
	}
	newClock := persist.FullClockHead(newPersistCommit.FullClock)

	diffs, err := d.getDiffsToMerge(fromCommits, toCommit)
	if err != nil {
		return err
	}

	_, err = d.getTerm(diffTable).Insert(diffs.Merge(func(diff gorethink.Term) map[string]interface{} {
		return map[string]interface{}{
			// the ID doesn't matter anymore, because the only reason why it had
			// to be a hash of (repo+commit+path) in PutFile is that we want
			// multiple clients writting the same file to be modifying the same
			// Diff.  But in Squash and Replay, the Diffs are not supposed to
			// be mutated anymore.
			"ID":    gorethink.UUID(),
			"Clock": newClock,
		}
	})).RunWrite(d.dbClient)
	if err != nil {
		return err
	}

	return nil
}

func (d *driver) ReplayCommit(fromCommits []*pfs.Commit, toBranch string) ([]*pfs.Commit, error) {
	if len(fromCommits) == 0 || toBranch == "" {
		return nil, fmt.Errorf("Invalid arguments: fromCommits: %v; toBranch: %v", fromCommits, toBranch)
	}

	var retCommits []*pfs.Commit
	// Make sure that all fromCommits are from the same repo
	var repo string
	for _, commit := range fromCommits {
		if repo == "" {
			repo = commit.Repo.Name
		} else {
			if repo != commit.Repo.Name {
				return nil, fmt.Errorf("you can only squash commits from the same repo; found %s and %s", repo, commit.Repo.Name)
			}
		}
	}

	commits, err := d.getCommitsToMerge(fromCommits, &pfs.Commit{
		Repo: &pfs.Repo{repo},
		ID:   toBranch,
	})
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
		newCommit, err := d.StartCommit(&pfs.Commit{
			Repo: &pfs.Repo{repo},
			ID:   toBranch,
		}, nil)
		if err != nil {
			return nil, err
		}

		rawCommitID, err := getRawCommitID(repo, newCommit.ID)
		if err != nil {
			return nil, err
		}

		cursor, err := d.getTerm(commitTable).Get(rawCommitID).Run(d.dbClient)
		var newPersistCommit persist.Commit
		if err := cursor.One(&newPersistCommit); err != nil {
			return nil, err
		}
		newClock := persist.FullClockHead(newPersistCommit.FullClock)
		oldClock := persist.FullClockHead(rawCommit.FullClock)

		// TODO: conflict detection
		_, err = d.getTerm(diffTable).Insert(d.getTerm(diffTable).GetAllByIndex(DiffClockIndex.Name, diffClockIndexKey(repo, oldClock.Branch, oldClock.Clock)).Merge(func(diff gorethink.Term) map[string]interface{} {
			return map[string]interface{}{
				"ID":    gorethink.UUID(),
				"Clock": newClock,
			}
		})).RunWrite(d.dbClient)
		if err != nil {
			return nil, err
		}

		err = d.FinishCommit(newCommit, false)
		retCommits = append(retCommits, newCommit)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
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
		fromClock, err = d.getFullClock(fromCommit)
		if err != nil {
			return nilTerm, err
		}
	}

	toClock, err := d.getFullClock(toCommit)
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

func (d *driver) getFullClock(to *pfs.Commit) (persist.FullClock, error) {
	commit, err := d.getRawCommit(to)
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

	return filterBlocks(diff, filterShard, file)
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
		if !pfsserver.FileInShard(filterShard, fileInfo.File) {
			continue
		}
		diff, err := filterBlocks(diff, filterShard, fileInfo.File)
		if err != nil {
			if _, ok := err.(*pfsserver.ErrFileNotFound); ok {
				continue
			}
			return nil, err
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
			ID:   diff.Clock.ReadableCommitID(),
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	return fileInfos, nil
}

func (d *driver) DeleteFile(file *pfs.File) error {
	fixPath(file)

	commit, err := d.getRawCommit(file.Commit)
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
			ID:        getDiffID(repo, commitID, path),
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

// getRawCommit accepts a repo name and an ID, and returns a Commit object.
// The ID can be of 2 forms:
// 1. branch/clock: like "master/3"
// 2. branch: like "master".  This would represent the head of the branch.
func (d *driver) getRawCommit(commit *pfs.Commit) (retCommit *persist.Commit, retErr error) {
	defer func() {
		if retErr == gorethink.ErrEmptyResult {
			retErr = pfsserver.NewErrCommitNotFound(commit.Repo.Name, commit.ID)
		}
	}()

	commitID, err := getRawCommitID(commit.Repo.Name, commit.ID)
	retCommit = &persist.Commit{}
	if err != nil {
		// We see if the commitID is a branch name
		if err := d.getHeadOfBranch(commit.Repo.Name, commit.ID, retCommit); err != nil {
			return nil, err
		}
	} else {
		cursor, err := d.getTerm(commitTable).Get(commitID).Run(d.dbClient)
		if err != nil {
			return nil, err
		}

		if err := cursor.One(retCommit); err != nil {
			return nil, err
		}
	}
	return retCommit, nil
}
