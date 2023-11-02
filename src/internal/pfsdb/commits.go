package pfsdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

const (
	// CommitsChannelName is used to watch events for the commits table.
	CommitsChannelName     = "pfs_commits"
	CommitsRepoChannelName = "pfs_commits_repo_"
	CommitChannelName      = "pfs_commits_"
	createCommit           = `
		WITH repo_row_id AS (SELECT id from pfs.repos WHERE name=$1 AND type=$2 AND project_id=(SELECT id from core.projects WHERE name=$3))
		INSERT INTO pfs.commits 
    	(commit_id, 
    	 commit_set_id, 
    	 repo_id, 
    	 branch_id, 
    	 description, 
    	 origin, 
    	 start_time, 
    	 finishing_time, 
    	 finished_time, 
    	 compacting_time_s, 
    	 validating_time_s, 
    	 size, 
    	 error) 
		VALUES 
		($4, $5,
		 (SELECT id from repo_row_id), 
		 (SELECT id from pfs.branches WHERE name=$6 AND repo_id=(SELECT id from repo_row_id)), 
		 $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING int_id;`
	updateCommit = `
		WITH repo_row_id AS (SELECT id from pfs.repos WHERE name=:repo.name AND type=:repo.type AND project_id=(SELECT id from core.projects WHERE name= :repo.project.name))
		UPDATE pfs.commits SET 
			commit_id=:commit_id, 
			commit_set_id=:commit_set_id,
		    repo_id=(SELECT id from repo_row_id), 
		    branch_id=(SELECT id from pfs.branches WHERE name=:branch_name AND repo_id=(SELECT id from repo_row_id)), 
			description=:description, 
			origin=:origin, 
			start_time=:start_time, 
			finishing_time=:finishing_time, 
			finished_time=:finished_time, 
		    compacting_time_s=:compacting_time_s, 
			validating_time_s=:validating_time_s, 
			size=:size, 
			error=:error 
		WHERE int_id=:int_id;`
	getCommit = `
		SELECT DISTINCT
    		commit.int_id, 
    		commit.commit_id, 
			commit.commit_set_id,
    		commit.branch_id, 
    		commit.origin, 
    		commit.description, 
    		commit.start_time, 
    		commit.finishing_time, 
    		commit.finished_time, 
    		commit.compacting_time_s, 
    		commit.validating_time_s,
    		commit.error, 
    		commit.size, 
    		commit.created_at,
    		commit.updated_at,
    		commit.repo_id AS "repo.id", 
    		repo.name AS "repo.name",
    		repo.type AS "repo.type",
    		project.name AS "repo.project.name",
    		branch.name as branch_name
		FROM pfs.commits commit
		JOIN pfs.repos repo ON commit.repo_id = repo.id
		JOIN core.projects project ON repo.project_id = project.id
		LEFT JOIN pfs.branches branch ON commit.branch_id = branch.id
		`
	getParentCommit = getCommit + `
		JOIN pfs.commit_ancestry ancestry ON ancestry.parent = commit.int_id`
	getChildCommit = getCommit + `
		JOIN pfs.commit_ancestry ancestry ON ancestry.child = commit.int_id`
)

// CommitNotFoundError is returned by GetCommit() when a commit is not found in postgres.
type CommitNotFoundError struct {
	RowID    CommitID
	CommitID string
}

func (err *CommitNotFoundError) Error() string {
	return fmt.Sprintf("commit (int_id=%d, commit_id=%s) not found", err.RowID, err.CommitID)
}

func (err *CommitNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

// ParentCommitNotFoundError is returned when a commit's parent is not found in postgres.
type ParentCommitNotFoundError struct {
	ChildRowID    CommitID
	ChildCommitID string
}

func IsParentCommitNotFound(err error) bool {
	return dbutil.IsNotNullViolation(err, "parent")
}

func (err *ParentCommitNotFoundError) Error() string {
	return fmt.Sprintf("parent commit of commit (int_id=%d, commit_id=%s) not found", err.ChildRowID, err.ChildCommitID)
}

func (err *ParentCommitNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

// ChildCommitNotFoundError is returned when a commit's child is not found in postgres.
type ChildCommitNotFoundError struct {
	Repo           string
	ParentRowID    CommitID
	ParentCommitID string
}

func IsChildCommitNotFound(err error) bool {
	return dbutil.IsNotNullViolation(err, "child")
}

func (err *ChildCommitNotFoundError) Error() string {
	return fmt.Sprintf("parent commit of commit (int_id=%d, commit_key=%s) not found", err.ParentRowID, err.ParentCommitID)
}

func (err *ChildCommitNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

// CommitMissingInfoError is returned when a commitInfo is missing a field.
type CommitMissingInfoError struct {
	Field string
}

func (err *CommitMissingInfoError) Error() string {
	return fmt.Sprintf("commitInfo.%s is missing/nil", err.Field)
}

func (err *CommitMissingInfoError) GRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, err.Error())
}

// CommitAlreadyExistsError is returned when a commit with the same name already exists in postgres.
type CommitAlreadyExistsError struct {
	CommitID string
}

// Error satisfies the error interface.
func (err *CommitAlreadyExistsError) Error() string {
	return fmt.Sprintf("commit %s already exists", err.CommitID)
}

func (err *CommitAlreadyExistsError) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, err.Error())
}

// AncestryOpt allows users to create commitInfos and skip creating the ancestry information.
// This allows a user to create the commits in an arbitrary order, then create their ancestry later.
type AncestryOpt struct {
	SkipChildren bool
	SkipParent   bool
}

// CommitsInRepoChannel returns the name of the channel that is notified when commits in repo 'repoID' are created, updated, or deleted
func CommitsInRepoChannel(repoID RepoID) string {
	return fmt.Sprintf("%s%d", CommitsRepoChannelName, repoID)
}

// CreateCommit creates an entry in the pfs.commits table. If the commit has a parent or children,
// it will attempt to create entries in the pfs.commit_ancestry table unless options are provided to skip
// ancestry creation.
func CreateCommit(ctx context.Context, tx *pachsql.Tx, commitInfo *pfs.CommitInfo, opts ...AncestryOpt) (CommitID, error) {
	if err := validateCommitInfo(commitInfo); err != nil {
		return 0, err
	}
	opt := AncestryOpt{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	branchName := sql.NullString{String: "", Valid: false}
	if commitInfo.Commit.Branch != nil {
		branchName = sql.NullString{String: commitInfo.Commit.Branch.Name, Valid: true}
	}
	insert := Commit{
		CommitID:    CommitKey(commitInfo.Commit),
		CommitSetID: commitInfo.Commit.Id,
		Repo: Repo{
			Name: commitInfo.Commit.Repo.Name,
			Type: commitInfo.Commit.Repo.Type,
			Project: Project{
				Name: commitInfo.Commit.Repo.Project.Name,
			},
		},
		BranchName:     branchName,
		Origin:         commitInfo.Origin.Kind.String(),
		StartTime:      pbutil.SanitizeTimestampPb(commitInfo.Started),
		FinishingTime:  pbutil.SanitizeTimestampPb(commitInfo.Finishing),
		FinishedTime:   pbutil.SanitizeTimestampPb(commitInfo.Finished),
		Description:    commitInfo.Description,
		CompactingTime: pbutil.DurationPbToBigInt(commitInfo.Details.CompactingTime),
		ValidatingTime: pbutil.DurationPbToBigInt(commitInfo.Details.ValidatingTime),
		Size:           commitInfo.Details.SizeBytes,
		Error:          commitInfo.Error,
	}
	// It would be nice to use a named query here, but sadly there is no NamedQueryRowContext. Additionally,
	// we run into errors when using named statements: (named statement already exists).
	row := tx.QueryRowxContext(ctx, createCommit, insert.Repo.Name, insert.Repo.Type, insert.Repo.Project.Name,
		insert.CommitID, insert.CommitSetID, insert.BranchName, insert.Description, insert.Origin, insert.StartTime, insert.FinishingTime,
		insert.FinishedTime, insert.CompactingTime, insert.ValidatingTime, insert.Size, insert.Error)
	if row.Err() != nil {
		if IsDuplicateKeyErr(row.Err()) { // a duplicate key implies that an entry for the repo already exists.
			return 0, &CommitAlreadyExistsError{CommitID: CommitKey(commitInfo.Commit)}
		}
		return 0, errors.Wrap(row.Err(), "exec create commitInfo")
	}
	lastInsertId := 0
	if err := row.Scan(&lastInsertId); err != nil {
		return 0, errors.Wrap(err, "scanning id from create commitInfo")
	}
	if commitInfo.ParentCommit != nil && !opt.SkipParent {
		if err := CreateCommitParent(ctx, tx, commitInfo.ParentCommit, CommitID(lastInsertId)); err != nil {
			return 0, errors.Wrap(err, "linking parent")
		}
	}
	if len(commitInfo.ChildCommits) != 0 && !opt.SkipChildren {
		if err := CreateCommitChildren(ctx, tx, CommitID(lastInsertId), commitInfo.ChildCommits); err != nil {
			return 0, errors.Wrap(err, "linking children")
		}
	}
	return CommitID(lastInsertId), nil
}

// CreateCommitParent inserts a single ancestry relationship where the child is known and parent must be derived.
func CreateCommitParent(ctx context.Context, tx *pachsql.Tx, parentCommit *pfs.Commit, childCommit CommitID) error {
	ancestryQuery := `
		INSERT INTO pfs.commit_ancestry
		(parent, child)
		VALUES ((SELECT int_id FROM pfs.commits WHERE commit_id=$1), $2)
		ON CONFLICT DO NOTHING;
	`
	_, err := tx.ExecContext(ctx, ancestryQuery, CommitKey(parentCommit), childCommit)
	if err != nil {
		if IsParentCommitNotFound(err) {
			return &ParentCommitNotFoundError{ChildRowID: childCommit}
		}
		return errors.Wrap(err, "putting commit parent")
	}
	return nil
}

// CreateCommitAncestries inserts ancestry relationships where the ids of both parent and children are known.
func CreateCommitAncestries(ctx context.Context, tx *pachsql.Tx, parentCommit CommitID, childrenCommits []CommitID) error {
	ancestryQueryTemplate := `
		INSERT INTO pfs.commit_ancestry
		(parent, child)
		VALUES %s
		ON CONFLICT DO NOTHING;
	`
	childValuesTemplate := `($1, $%d)`
	params := []any{parentCommit}
	queryVarNum := 2
	values := make([]string, 0)
	for _, child := range childrenCommits {
		values = append(values, fmt.Sprintf(childValuesTemplate, queryVarNum))
		params = append(params, child)
		queryVarNum++
	}
	_, err := tx.ExecContext(ctx, fmt.Sprintf(ancestryQueryTemplate, strings.Join(values, ",")),
		params...)
	if err != nil {
		if IsChildCommitNotFound(err) {
			return &ChildCommitNotFoundError{ParentRowID: parentCommit}
		}
		return errors.Wrap(err, "putting commit children")
	}
	return nil
}

// CreateCommitChildren inserts ancestry relationships using a single query for all of the children.
func CreateCommitChildren(ctx context.Context, tx *pachsql.Tx, parentCommit CommitID, childCommits []*pfs.Commit) error {
	ancestryQueryTemplate := `
		INSERT INTO pfs.commit_ancestry
		(parent, child)
		VALUES %s
		ON CONFLICT DO NOTHING;
	`
	childValuesTemplate := `($1, (SELECT int_id FROM pfs.commits WHERE commit_id=$%d))`
	params := []any{parentCommit}
	queryVarNum := 2
	values := make([]string, 0)
	for _, child := range childCommits {
		values = append(values, fmt.Sprintf(childValuesTemplate, queryVarNum))
		params = append(params, CommitKey(child))
		queryVarNum++
	}
	_, err := tx.ExecContext(ctx, fmt.Sprintf(ancestryQueryTemplate, strings.Join(values, ",")),
		params...)
	if err != nil {
		if IsChildCommitNotFound(err) {
			return &ChildCommitNotFoundError{ParentRowID: parentCommit}
		}
		return errors.Wrap(err, "putting commit children")
	}
	return nil
}

// DeleteCommit deletes an entry in the pfs.commits table. It also repoints the references in the commit_ancestry table.
// The caller is responsible for updating branchesg.
func DeleteCommit(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) error {
	if commit == nil {
		return &CommitMissingInfoError{Field: "Commit"}
	}
	id, err := GetCommitID(ctx, tx, commit)
	if err != nil {
		return errors.Wrap(err, "getting commit ID to delete")
	}
	parent, children, err := getCommitRelativeRows(ctx, tx, id)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("getting commit relatives for id=%d", id))
	}
	// delete commit.parent -> commit and commit -> commit.children if they exist.
	if parent != nil || children != nil {
		_, err = tx.ExecContext(ctx, "DELETE FROM pfs.commit_ancestry WHERE parent=$1 OR child=$1;", id)
		if err != nil {
			return errors.Wrap(err, "delete commit ancestry")
		}
	}
	// repoint commit.parent -> commit.children
	if parent != nil && children != nil {
		childrenIDs := make([]CommitID, 0)
		for _, child := range children {
			childrenIDs = append(childrenIDs, child.ID)
		}
		if err := CreateCommitAncestries(ctx, tx, parent.ID, childrenIDs); err != nil {
			return errors.Wrap(err, fmt.Sprintf("repointing id=%d at %v", parent.ID, childrenIDs))
		}
	}
	// delete commit.
	result, err := tx.ExecContext(ctx, "DELETE FROM pfs.commits WHERE int_id=$1;", id)
	if err != nil {
		return errors.Wrap(err, "delete commit")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "could not get affected rows")
	}
	if rowsAffected == 0 {
		return &CommitNotFoundError{RowID: id}
	}
	return nil
}

// GetCommitID returns the int_id of a commit in postgres.
func GetCommitID(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) (CommitID, error) {
	if commit == nil {
		return 0, &CommitMissingInfoError{Field: "Commit"}
	}
	if commit.Repo == nil {
		return 0, &CommitMissingInfoError{Field: "Repo"}
	}
	row, err := getCommitRowByCommitKey(ctx, tx, commit)
	if err != nil {
		return 0, errors.Wrap(err, "get commit by commit key")
	}
	return row.ID, nil
}

// GetCommit returns the commitInfo where int_id=id.
func GetCommit(ctx context.Context, tx *pachsql.Tx, id CommitID) (*pfs.CommitInfo, error) {
	if id == 0 {
		return nil, errors.New("invalid id: 0")
	}
	row := &Commit{}
	err := tx.QueryRowxContext(ctx, fmt.Sprintf("%s WHERE int_id=$1", getCommit), id).StructScan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &CommitNotFoundError{RowID: id}
		}
		return nil, errors.Wrap(err, "scanning commit row")
	}
	commitInfo, err := getCommitInfoFromCommitRow(ctx, tx, row)
	if err != nil {
		return nil, errors.Wrap(err, "get commit info from row")
	}
	return commitInfo, err
}

func GetCommitWithIDByKey(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) (*CommitWithID, error) {
	row, err := getCommitRowByCommitKey(ctx, tx, commit)
	if err != nil {
		return nil, errors.Wrap(err, "get commit by commit key")
	}
	commitInfo, err := getCommitInfoFromCommitRow(ctx, tx, row)
	if err != nil {
		return nil, errors.Wrap(err, "get commit info from row")
	}
	return &CommitWithID{
		CommitInfo: commitInfo,
		ID:         row.ID,
	}, nil
}

// GetCommitByCommitKey is like GetCommit but derives the int_id on behalf of the caller.
func GetCommitByCommitKey(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) (*pfs.CommitInfo, error) {
	pair, err := GetCommitWithIDByKey(ctx, tx, commit)
	if err != nil {
		return nil, err
	}
	return pair.CommitInfo, nil
}

// GetCommitParent uses the pfs.commit_ancestry and pfs.commits tables to retrieve a commit given an int_id of
// one of its children.
func GetCommitParent(ctx context.Context, tx *pachsql.Tx, childCommit CommitID) (*pfs.Commit, error) {
	return getCommitParent(ctx, tx, childCommit)
}

// GetCommitParent uses the pfs.commit_ancestry and pfs.commits tables to retrieve a commit given an int_id of
// one of its children.
func getCommitParent(ctx context.Context, extCtx sqlx.ExtContext, childCommit CommitID) (*pfs.Commit, error) {
	row, err := getCommitParentRow(ctx, extCtx, childCommit)
	if err != nil {
		return nil, errors.Wrap(err, "getting parent commit row")
	}
	parentCommitInfo := parseCommitInfoFromRow(row)
	return parentCommitInfo.Commit, nil
}

// GetCommitChildren uses the pfs.commit_ancestry and pfs.commits tables to retrieve commits of all
// of the children given an int_id of the parent.
func GetCommitChildren(ctx context.Context, tx *pachsql.Tx, parentCommit CommitID) ([]*pfs.Commit, error) {
	return getCommitChildren(ctx, tx, parentCommit)
}

func getCommitChildren(ctx context.Context, extCtx sqlx.ExtContext, parentCommit CommitID) ([]*pfs.Commit, error) {
	children := make([]*pfs.Commit, 0)
	rows, err := extCtx.QueryxContext(ctx, fmt.Sprintf("%s WHERE ancestry.parent=$1", getChildCommit), parentCommit)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ChildCommitNotFoundError{ParentRowID: parentCommit}
		}
		return nil, errors.Wrap(err, "getting commit children")
	}
	for rows.Next() {
		row := &Commit{}
		if err := rows.StructScan(row); err != nil {
			return nil, errors.Wrap(err, "scanning commit row for child")
		}
		childCommitInfo := parseCommitInfoFromRow(row)
		children = append(children, childCommitInfo.Commit)
	}
	if len(children) == 0 { // QueryxContext does not return an error when the query returns 0 rows.
		return nil, &ChildCommitNotFoundError{ParentRowID: parentCommit}
	}
	return children, nil
}

// UpsertCommit will attempt to insert a commit and its ancestry relationships.
// If the commit already exists, it will update its description.
func UpsertCommit(ctx context.Context, tx *pachsql.Tx, commitInfo *pfs.CommitInfo, opts ...AncestryOpt) (CommitID, error) {
	existingCommit, err := getCommitRowByCommitKey(ctx, tx, commitInfo.Commit)
	if err != nil {
		if errors.As(err, &CommitNotFoundError{CommitID: CommitKey(commitInfo.Commit)}) {
			return CreateCommit(ctx, tx, commitInfo, opts...)
		}
		return 0, errors.Wrap(err, "upserting commit")
	}
	return existingCommit.ID, UpdateCommit(ctx, tx, existingCommit.ID, commitInfo, opts...)
}

// UpdateCommit overwrites an existing commit entry by CommitID as well as the corresponding ancestry entries.
func UpdateCommit(ctx context.Context, tx *pachsql.Tx, id CommitID, commitInfo *pfs.CommitInfo, opts ...AncestryOpt) error {
	if err := validateCommitInfo(commitInfo); err != nil {
		return err
	}
	opt := AncestryOpt{}
	if len(opts) > 0 {
		opt = opts[0]
	}
	branchName := sql.NullString{String: "", Valid: false}
	if commitInfo.Commit.Branch != nil {
		branchName = sql.NullString{String: commitInfo.Commit.Branch.Name, Valid: true}
	}
	update := Commit{
		ID:          id,
		CommitID:    CommitKey(commitInfo.Commit),
		CommitSetID: commitInfo.Commit.Id,
		Repo: Repo{
			Name: commitInfo.Commit.Repo.Name,
			Type: commitInfo.Commit.Repo.Type,
			Project: Project{
				Name: commitInfo.Commit.Repo.Project.Name,
			},
		},
		BranchName:     branchName,
		Origin:         commitInfo.Origin.Kind.String(),
		StartTime:      pbutil.SanitizeTimestampPb(commitInfo.Started),
		FinishingTime:  pbutil.SanitizeTimestampPb(commitInfo.Finishing),
		FinishedTime:   pbutil.SanitizeTimestampPb(commitInfo.Finished),
		Description:    commitInfo.Description,
		CompactingTime: pbutil.DurationPbToBigInt(commitInfo.Details.CompactingTime),
		ValidatingTime: pbutil.DurationPbToBigInt(commitInfo.Details.ValidatingTime),
		Size:           commitInfo.Details.SizeBytes,
		Error:          commitInfo.Error,
	}
	query := updateCommit
	res, err := tx.NamedExecContext(ctx, query, update)
	if err != nil {
		return errors.Wrap(err, "exec update commitInfo")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "rows affected")
	}
	if rowsAffected == 0 {
		_, err := GetRepoByName(ctx, tx, commitInfo.Commit.Repo.Project.Name, commitInfo.Commit.Repo.Name, commitInfo.Commit.Repo.Type)
		if err != nil {
			return err
		}
		return &CommitNotFoundError{RowID: id}
	}
	if !opt.SkipParent {
		_, err = tx.ExecContext(ctx, "DELETE FROM pfs.commit_ancestry WHERE child=$1;", id)
		if err != nil {
			return errors.Wrap(err, "delete commit parent")
		}
		if commitInfo.ParentCommit != nil {
			if err := CreateCommitParent(ctx, tx, commitInfo.ParentCommit, id); err != nil {
				return errors.Wrap(err, "linking parent")
			}
		}
	}
	if opt.SkipChildren {
		return nil
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM pfs.commit_ancestry WHERE parent=$1;", id)
	if err != nil {
		return errors.Wrap(err, "delete commit children")
	}
	if len(commitInfo.ChildCommits) != 0 {
		if err := CreateCommitChildren(ctx, tx, id, commitInfo.ChildCommits); err != nil {
			return errors.Wrap(err, "linking children")
		}
	}
	return nil
}

// validateCommitInfo returns an error if the commit is not valid and has side effects of instantiating details
// if they are nil.
func validateCommitInfo(commitInfo *pfs.CommitInfo) error {
	if commitInfo.Commit == nil {
		return &CommitMissingInfoError{Field: "Commit"}
	}
	if commitInfo.Commit.Repo == nil {
		return &CommitMissingInfoError{Field: "Repo"}
	}
	if commitInfo.Origin == nil {
		return &CommitMissingInfoError{Field: "Origin"}
	}
	if commitInfo.Details == nil { // stub in an empty details struct to avoid panics.
		commitInfo.Details = &pfs.CommitInfo_Details{}
	}
	switch commitInfo.Origin.Kind {
	case pfs.OriginKind_ORIGIN_KIND_UNKNOWN, pfs.OriginKind_USER, pfs.OriginKind_AUTO, pfs.OriginKind_FSCK:
		break
	default:
		return errors.New(fmt.Sprintf("invalid origin: %v", commitInfo.Origin.Kind))
	}
	return nil
}

func getCommitInfoFromCommitRow(ctx context.Context, extCtx sqlx.ExtContext, row *Commit) (*pfs.CommitInfo, error) {
	var err error
	commitInfo := parseCommitInfoFromRow(row)
	commitInfo.ParentCommit, commitInfo.ChildCommits, err = getCommitRelatives(ctx, extCtx, row.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get commit relatives")
	}
	commitInfo.DirectProvenance, err = CommitDirectProvenance(ctx, extCtx, row.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get provenance for commit")
	}
	return commitInfo, err
}

func getCommitRelatives(ctx context.Context, extCtx sqlx.ExtContext, commitID CommitID) (*pfs.Commit, []*pfs.Commit, error) {
	parentCommit, err := getCommitParent(ctx, extCtx, commitID)
	if err != nil && !errors.As(err, &ParentCommitNotFoundError{ChildRowID: commitID}) {
		return nil, nil, errors.Wrap(err, "getting parent commit")
		// if parent is missing, assume commit is root of a repo.
	}
	childCommits, err := getCommitChildren(ctx, extCtx, commitID)
	if err != nil && !errors.As(err, &ChildCommitNotFoundError{ParentRowID: commitID}) {
		return nil, nil, errors.Wrap(err, "getting children commits")
		// if children is missing, assume commit is HEAD of some branch.
	}
	return parentCommit, childCommits, nil
}

func getCommitParentRow(ctx context.Context, extCtx sqlx.ExtContext, childCommit CommitID) (*Commit, error) {
	row := &Commit{}
	if err := sqlx.GetContext(ctx, extCtx, row, fmt.Sprintf("%s WHERE ancestry.child=$1", getParentCommit), childCommit); err != nil {
		if err == sql.ErrNoRows {
			return nil, &ParentCommitNotFoundError{ChildRowID: childCommit}
		}
		return nil, errors.Wrap(err, "scanning commit row for parent")
	}
	return row, nil
}

func getCommitChildrenRows(ctx context.Context, tx *pachsql.Tx, parentCommit CommitID) ([]*Commit, error) {
	children := make([]*Commit, 0)
	rows, err := tx.QueryxContext(ctx, fmt.Sprintf("%s WHERE ancestry.parent=$1", getChildCommit), parentCommit)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &ChildCommitNotFoundError{ParentRowID: parentCommit}
		}
		return nil, errors.Wrap(err, "getting commit children rows")
	}
	for rows.Next() {
		row := &Commit{}
		if err := rows.StructScan(row); err != nil {
			return nil, errors.Wrap(err, "scanning commit row for child")
		}
		children = append(children, row)
	}
	if len(children) == 0 { // QueryxContext does not return an error when the query returns 0 rows.
		return nil, &ChildCommitNotFoundError{ParentRowID: parentCommit}
	}
	return children, nil
}

func getCommitRelativeRows(ctx context.Context, tx *pachsql.Tx, commitID CommitID) (*Commit, []*Commit, error) {
	commitParentRows, err := getCommitParentRow(ctx, tx, commitID)
	if err != nil && !errors.As(err, &ParentCommitNotFoundError{ChildRowID: commitID}) {
		return nil, nil, errors.Wrap(err, "getting parent commit")
		// if parent is missing, assume commit is root of a repo.
	}
	commitChildrenRows, err := getCommitChildrenRows(ctx, tx, commitID)
	if err != nil && !errors.As(err, &ChildCommitNotFoundError{ParentRowID: commitID}) {
		return nil, nil, errors.Wrap(err, "getting children commits")
		// if children is missing, assume commit is HEAD of some branch.
	}
	return commitParentRows, commitChildrenRows, nil
}

func getCommitRowByCommitKey(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) (*Commit, error) {
	row := &Commit{}
	if commit == nil {
		return nil, &CommitMissingInfoError{Field: "Commit"}
	}
	id := CommitKey(commit)
	err := tx.QueryRowxContext(ctx, fmt.Sprintf("%s WHERE commit_id=$1", getCommit), id).StructScan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err := GetRepoByName(ctx, tx, commit.Repo.Project.Name, commit.Repo.Name, commit.Repo.Type)
			if err != nil {
				return nil, err
			}
			return nil, &CommitNotFoundError{CommitID: id}
		}
		return nil, errors.Wrap(err, "scanning commit row")
	}
	return row, nil
}

func parseCommitInfoFromRow(row *Commit) *pfs.CommitInfo {
	commitInfo := &pfs.CommitInfo{
		Commit:      row.Pb(),
		Origin:      &pfs.CommitOrigin{Kind: pfs.OriginKind(pfs.OriginKind_value[strings.ToUpper(row.Origin)])},
		Started:     pbutil.TimeToTimestamppb(row.StartTime),
		Finishing:   pbutil.TimeToTimestamppb(row.FinishingTime),
		Finished:    pbutil.TimeToTimestamppb(row.FinishedTime),
		Description: row.Description,
		Error:       row.Error,
		Details: &pfs.CommitInfo_Details{
			CompactingTime: pbutil.BigIntToDurationpb(row.CompactingTime),
			ValidatingTime: pbutil.BigIntToDurationpb(row.ValidatingTime),
			SizeBytes:      row.Size,
		},
	}
	return commitInfo
}

// CommitWithID is returned by the commit iterator.
type CommitWithID struct {
	ID         CommitID
	CommitInfo *pfs.CommitInfo
	Revision   int64
}

// this dropped global variable instantiation forces the compiler to check whether CommitIterator implements stream.Iterator.
var _ stream.Iterator[CommitWithID] = &CommitIterator{}

// commitColumn is used in the ListCommit filter and defines specific field names for type safety.
// This should hopefully prevent a library user from misconfiguring the filter.
type commitColumn string

var (
	CommitColumnID        = commitColumn("commit.int_id")
	CommitColumnSetID     = commitColumn("commit.commit_set_id")
	CommitColumnOrigin    = commitColumn("commit.origin")
	CommitColumnCreatedAt = commitColumn("commit.created_at")
	CommitColumnUpdatedAt = commitColumn("commit.updated_at")
)

type OrderByCommitColumn OrderByColumn[commitColumn]

type CommitIterator struct {
	paginator pageIterator[Commit]
	extCtx    sqlx.ExtContext
}

func (i *CommitIterator) Next(ctx context.Context, dst *CommitWithID) error {
	if dst == nil {
		return errors.Errorf("dst CommitInfo cannot be nil")
	}
	commit, rev, err := i.paginator.next(ctx, i.extCtx)
	if err != nil {
		return err
	}
	commitInfo, err := getCommitInfoFromCommitRow(ctx, i.extCtx, commit)
	if err != nil {
		return err
	}
	dst.ID = commit.ID
	dst.CommitInfo = commitInfo
	dst.Revision = rev
	return nil
}

func NewCommitsIterator(ctx context.Context, extCtx sqlx.ExtContext, startPage, pageSize uint64, filter *pfs.Commit, orderBys ...OrderByCommitColumn) (*CommitIterator, error) {
	var conditions []string
	var values []any
	// Note that using ? as the bindvar is okay because we rebind it later.
	if filter != nil {
		if filter.Repo != nil && filter.Repo.Name != "" {
			conditions = append(conditions, "repo.name = ?")
			values = append(values, filter.Repo.Name)
		}
		if filter.Repo != nil && filter.Repo.Type != "" {
			conditions = append(conditions, "repo.type = ?")
			values = append(values, filter.Repo.Type)
		}
		if filter.Repo != nil && filter.Repo.Project != nil && filter.Repo.Project.Name != "" {
			conditions = append(conditions, "project.name = ?")
			values = append(values, filter.Repo.Project.Name)
		}
		if filter.Id != "" {
			conditions = append(conditions, "commit.commit_set_id = ?")
			values = append(values, filter.Id)
		}
		if filter.Branch != nil && filter.Branch.Name != "" {
			conditions = append(conditions, "branch.name = ?")
			values = append(values, filter.Branch.Name)
		}
	}
	query := getCommit
	if len(conditions) > 0 {
		query += fmt.Sprintf("\nWHERE %s\n", strings.Join(conditions, " AND "))
	}
	// Compute ORDER BY
	var orderByGeneric []OrderByColumn[commitColumn]
	if len(orderBys) == 0 {
		orderByGeneric = []OrderByColumn[commitColumn]{{Column: CommitColumnID, Order: SortOrderAsc}}
	} else {
		for _, orderBy := range orderBys {
			orderByGeneric = append(orderByGeneric, OrderByColumn[commitColumn](orderBy))
		}
	}
	query = extCtx.Rebind(query + OrderByQuery[commitColumn](orderByGeneric...))
	return &CommitIterator{
		paginator: newPageIterator[Commit](ctx, query, values, startPage, pageSize),
		extCtx:    extCtx,
	}, nil
}

// ListCommit returns a CommitIterator that exposes a Next() function for retrieving *pfs.CommitInfo references.
// It manages transactions on behalf of its user under the hood.
func ListCommit(ctx context.Context, db *pachsql.DB, filter *pfs.Commit, orderBys ...OrderByCommitColumn) (*CommitIterator, error) {
	return NewCommitsIterator(ctx, db, 0, 100, filter, orderBys...)
}

func UpdateCommitTxByFilter(ctx context.Context, tx *pachsql.Tx, filter *pfs.Commit, cb func(commitWithID CommitWithID) error, orderBys ...OrderByCommitColumn) error {
	return errors.Wrap(listCommitTxByFilter(ctx, tx, filter, func(commitWithID CommitWithID) error {
		if err := cb(commitWithID); err != nil {
			return err
		}
		return UpdateCommit(ctx, tx, commitWithID.ID, commitWithID.CommitInfo)
	}, orderBys...), "update commits tx by filter")
}

func ListCommitTxByFilter(ctx context.Context, tx *pachsql.Tx, filter *pfs.Commit, orderBys ...OrderByCommitColumn) ([]*pfs.CommitInfo, error) {
	var commits []*pfs.CommitInfo
	if err := listCommitTxByFilter(ctx, tx, filter, func(commitWithID CommitWithID) error {
		commits = append(commits, commitWithID.CommitInfo)
		return nil
	}, orderBys...); err != nil {
		return nil, errors.Wrap(err, "list commits tx by filter")
	}
	return commits, nil
}

func listCommitTxByFilter(ctx context.Context, tx *pachsql.Tx, filter *pfs.Commit, cb func(commitWithID CommitWithID) error, orderBys ...OrderByCommitColumn) error {
	if filter == nil {
		return errors.Errorf("filter cannot be empty")
	}
	iter, err := NewCommitsIterator(ctx, tx, 0, 100, filter, orderBys...)
	if err != nil {
		return errors.Wrap(err, "list commit tx by filter")
	}
	if err := stream.ForEach[CommitWithID](ctx, iter, func(commitWithID CommitWithID) error {
		return cb(commitWithID)
	}); err != nil {
		return errors.Wrap(err, "list index")
	}
	return nil
}
