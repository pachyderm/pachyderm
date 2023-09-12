package pfsdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// CommitsChannelName is used to watch events for the commits table.
	CommitsChannelName = "pfs_commits"
	createCommit       = `
		WITH repo_row_id AS (SELECT id from pfs.repos WHERE name=:repo.name AND type=:repo.type AND project_id=(SELECT id from core.projects WHERE name= :repo.project.name))
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
		(:commit_id, 
		 :commit_set_id, 
		 (SELECT id from repo_row_id), 
		(SELECT id from pfs.branches WHERE name=:branch_name AND repo_id=(SELECT id from repo_row_id)), 
		:description, 
		:origin, 
		:start_time,
		:finishing_time,
		:finished_time,
		:compacting_time_s,
		:validating_time_s,
		:size,
		:error)
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
		SELECT 
    		commit.int_id, 
    		commit.commit_id, 
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
    		commit.repo_id AS "repo.id", 
    		repo.name AS "repo.name",
    		repo.type AS "repo.type",
    		project.name AS "repo.project.name",
    		branch.name as branch_name
		FROM pfs.commits commit
		JOIN pfs.repos repo ON commit.repo_id = repo.id
		JOIN core.projects project ON repo.project_id = project.id
		LEFT JOIN pfs.branches branch ON commit.branch_id = branch.id`
	getParentCommit = getCommit + `
		JOIN pfs.commit_ancestry ancestry ON ancestry.parent = commit.int_id`
	getChildCommit = getCommit + `
		JOIN pfs.commit_ancestry ancestry ON ancestry.child = commit.int_id`
)

// ErrCommitNotFound is returned by GetCommit() when a commit is not found in postgres.
type ErrCommitNotFound struct {
	RowID    CommitID
	CommitID string
}

func (err ErrCommitNotFound) Error() string {
	return fmt.Sprintf("commit (int_id=%d, commit_id=%s) not found", err.RowID, err.CommitID)
}

func (err ErrCommitNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

// ErrParentCommitNotFound is returned when a commit's parent is not found in postgres.
type ErrParentCommitNotFound struct {
	ChildRowID    CommitID
	ChildCommitID string
}

func IsParentCommitNotFound(err error) bool {
	return dbutil.IsNotNullViolation(err, "parent")
}

func (err ErrParentCommitNotFound) Error() string {
	return fmt.Sprintf("parent commit of commit (int_id=%d, commit_id=%s) not found", err.ChildRowID, err.ChildCommitID)
}

func (err ErrParentCommitNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

// ErrChildCommitNotFound is returned when a commit's child is not found in postgres.
type ErrChildCommitNotFound struct {
	Repo           string
	ParentRowID    CommitID
	ParentCommitID string
}

func IsChildCommitNotFound(err error) bool {
	return dbutil.IsNotNullViolation(err, "child")
}

func (err ErrChildCommitNotFound) Error() string {
	return fmt.Sprintf("parent commit of commit (int_id=%d, commit_key=%s) not found", err.ParentRowID, err.ParentCommitID)
}

func (err ErrChildCommitNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

// ErrCommitMissingInfo is returned when a commitInfo is missing a field.
type ErrCommitMissingInfo struct {
	Field string
}

func (err ErrCommitMissingInfo) Error() string {
	return fmt.Sprintf("commitInfo.%s is missing/nil", err.Field)
}

func (err ErrCommitMissingInfo) GRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, err.Error())
}

// ErrCommitAlreadyExists is returned when a commit with the same name already exists in postgres.
type ErrCommitAlreadyExists struct {
	CommitID string
}

// Error satisfies the error interface.
func (err ErrCommitAlreadyExists) Error() string {
	return fmt.Sprintf("commit %s already exists", err.CommitID)
}

func (err ErrCommitAlreadyExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, err.Error())
}

// AncestryOpt allows users to create commitInfos and skip creating the ancestry information.
// This allows a user to create the commits in an arbitrary order, then create their ancestry later.
type AncestryOpt struct {
	SkipChildren bool
	SkipParent   bool
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
	insert := Commit{
		CommitID:    CommitKey(commitInfo.Commit),
		CommitSetID: commitInfo.Commit.Id,
		Repo: Repo{
			Name: commitInfo.Commit.Repo.Name,
			Type: commitInfo.Commit.Repo.Type,
			Project: coredb.Project{
				Name: commitInfo.Commit.Repo.Project.Name,
			},
		},
		BranchName:     sql.NullString{String: commitInfo.Commit.Branch.Name, Valid: true},
		Origin:         commitInfo.Origin.Kind.String(),
		StartTime:      sanitizeTimestamppb(commitInfo.Started),
		FinishingTime:  sanitizeTimestamppb(commitInfo.Finishing),
		FinishedTime:   sanitizeTimestamppb(commitInfo.Finished),
		Description:    commitInfo.Description,
		CompactingTime: durationpbToBigInt(commitInfo.Details.CompactingTime),
		ValidatingTime: durationpbToBigInt(commitInfo.Details.ValidatingTime),
		Size:           commitInfo.Details.SizeBytes,
		Error:          commitInfo.Error,
	}
	namedStmt, err := tx.PrepareNamedContext(ctx, createCommit)
	if err != nil {
		return 0, errors.Wrap(err, "prepare create commitInfo")
	}
	row := namedStmt.QueryRowxContext(ctx, insert)
	if row.Err() != nil {
		if IsDuplicateKeyErr(row.Err()) { // a duplicate key implies that an entry for the repo already exists.
			return 0, ErrCommitAlreadyExists{CommitID: CommitKey(commitInfo.Commit)}
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
			return ErrParentCommitNotFound{ChildRowID: childCommit}
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
			return ErrChildCommitNotFound{ParentRowID: parentCommit}
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
			return ErrChildCommitNotFound{ParentRowID: parentCommit}
		}
		return errors.Wrap(err, "putting commit children")
	}
	return nil
}

// DeleteCommit deletes an entry in the pfs.commits table. It also repoints the references in the commit_ancestry table.
// The caller is responsible for updating branchesg.
func DeleteCommit(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) error {
	if commit == nil {
		return ErrCommitMissingInfo{Field: "Commit"}
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
		commitsNotFinished := make([]string, 0)
		for _, child := range children {
			if child.FinishedTime.IsZero() {
				commitsNotFinished = append(commitsNotFinished, child.CommitID)
			}
			childrenIDs = append(childrenIDs, child.ID)
		}
		if len(commitsNotFinished) > 0 {
			return errors.New(fmt.Sprintf("commits not finished before deleting: %v", commitsNotFinished))
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
		return ErrCommitNotFound{RowID: id}
	}
	return nil
}

// GetCommitID returns the int_id of a commit in postgres.
func GetCommitID(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) (CommitID, error) {
	if commit == nil {
		return 0, ErrCommitMissingInfo{Field: "Commit"}
	}
	if commit.Repo == nil {
		return 0, ErrCommitMissingInfo{Field: "Repo"}
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
			return nil, ErrCommitNotFound{RowID: id}
		}
		return nil, errors.Wrap(err, "scanning commit row")
	}
	commitInfo, err := getCommitInfoFromCommitRow(ctx, tx, row)
	if err != nil {
		return nil, errors.Wrap(err, "get commit info from row")
	}
	return commitInfo, err
}

// GetCommitByCommitKey is like GetCommit but derives the int_id on behalf of the caller.
func GetCommitByCommitKey(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) (*pfs.CommitInfo, error) {
	row, err := getCommitRowByCommitKey(ctx, tx, commit)
	if err != nil {
		return nil, errors.Wrap(err, "get commit by commit key")
	}
	commitInfo, err := getCommitInfoFromCommitRow(ctx, tx, row)
	if err != nil {
		return nil, errors.Wrap(err, "get commit info from row")
	}
	return commitInfo, err
}

// GetCommitParent uses the pfs.commit_ancestry and pfs.commits tables to retrieve a commit given an int_id of
// one of its children.
func GetCommitParent(ctx context.Context, tx *pachsql.Tx, childCommit CommitID) (*pfs.Commit, error) {
	row, err := getCommitParentRow(ctx, tx, childCommit)
	if err != nil {
		return nil, errors.Wrap(err, "getting parent commit row")
	}
	parentCommitInfo, err := parseCommitInfoFromRow(row)
	if err != nil {
		return nil, errors.Wrap(err, "get parent commit from row")
	}
	return parentCommitInfo.Commit, nil
}

// GetCommitChildren uses the pfs.commit_ancestry and pfs.commits tables to retrieve commits of all
// of the children given an int_id of the parent.
func GetCommitChildren(ctx context.Context, tx *pachsql.Tx, parentCommit CommitID) ([]*pfs.Commit, error) {
	children := make([]*pfs.Commit, 0)
	rows, err := tx.QueryxContext(ctx, fmt.Sprintf("%s WHERE ancestry.parent=$1", getChildCommit), parentCommit)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrChildCommitNotFound{ParentRowID: parentCommit}
		}
		return nil, errors.Wrap(err, "getting commit children")
	}
	for rows.Next() {
		row := &Commit{}
		if err := rows.StructScan(row); err != nil {
			return nil, errors.Wrap(err, "scanning commit row for child")
		}
		childCommitInfo, err := parseCommitInfoFromRow(row)
		if err != nil {
			return nil, errors.Wrap(err, "getting child commit from row")
		}
		children = append(children, childCommitInfo.Commit)
	}
	if len(children) == 0 { // QueryxContext does not return an error when the query returns 0 rows.
		return nil, ErrChildCommitNotFound{ParentRowID: parentCommit}
	}
	return children, nil
}

// UpsertCommit will attempt to insert a commit and its ancestry relationships.
// If the commit already exists, it will update its description.
func UpsertCommit(ctx context.Context, tx *pachsql.Tx, commitInfo *pfs.CommitInfo, opts ...AncestryOpt) (CommitID, error) {
	existingCommit, err := getCommitRowByCommitKey(ctx, tx, commitInfo.Commit)
	if err != nil {
		if errors.Is(ErrCommitNotFound{CommitID: CommitKey(commitInfo.Commit)}, errors.Cause(err)) {
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
	update := Commit{
		ID:          id,
		CommitID:    CommitKey(commitInfo.Commit),
		CommitSetID: commitInfo.Commit.Id,
		Repo: Repo{
			Name: commitInfo.Commit.Repo.Name,
			Type: commitInfo.Commit.Repo.Type,
			Project: coredb.Project{
				Name: commitInfo.Commit.Repo.Project.Name,
			},
		},
		BranchName:     sql.NullString{String: commitInfo.Commit.Branch.Name, Valid: true},
		Origin:         commitInfo.Origin.Kind.String(),
		StartTime:      sanitizeTimestamppb(commitInfo.Started),
		FinishingTime:  sanitizeTimestamppb(commitInfo.Finishing),
		FinishedTime:   sanitizeTimestamppb(commitInfo.Finished),
		Description:    commitInfo.Description,
		CompactingTime: durationpbToBigInt(commitInfo.Details.CompactingTime),
		ValidatingTime: durationpbToBigInt(commitInfo.Details.ValidatingTime),
		Size:           commitInfo.Details.SizeBytes,
		Error:          commitInfo.Error,
	}
	query := updateCommit
	_, err := tx.NamedExecContext(ctx, query, update)
	if err != nil {
		return errors.Wrap(err, "exec update commitInfo")
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
		return ErrCommitMissingInfo{Field: "Commit"}
	}
	if commitInfo.Commit.Repo == nil {
		return ErrCommitMissingInfo{Field: "Repo"}
	}
	if commitInfo.Origin == nil {
		return ErrCommitMissingInfo{Field: "Origin"}
	}
	if commitInfo.Commit.Branch == nil {
		return ErrCommitMissingInfo{Field: "Branch"}
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

func getCommitInfoFromCommitRow(ctx context.Context, tx *pachsql.Tx, row *Commit) (*pfs.CommitInfo, error) {
	commitInfo, err := parseCommitInfoFromRow(row)
	if err != nil {
		return nil, errors.Wrap(err, "get commit from row")
	}
	commitInfo.ParentCommit, commitInfo.ChildCommits, err = getCommitRelatives(ctx, tx, row.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get commit relatives")
	}
	return commitInfo, err
}

func getCommitRelatives(ctx context.Context, tx *pachsql.Tx, commitID CommitID) (*pfs.Commit, []*pfs.Commit, error) {
	parentCommit, err := GetCommitParent(ctx, tx, commitID)
	if err != nil && !errors.Is(ErrParentCommitNotFound{ChildRowID: commitID}, errors.Cause(err)) {
		return nil, nil, errors.Wrap(err, "getting parent commit")
		// if parent is missing, assume commit is root of a repo.
	}
	childCommits, err := GetCommitChildren(ctx, tx, commitID)
	if err != nil && !errors.Is(ErrChildCommitNotFound{ParentRowID: commitID}, errors.Cause(err)) {
		return nil, nil, errors.Wrap(err, "getting children commits")
		// if children is missing, assume commit is HEAD of some branch.
	}
	return parentCommit, childCommits, nil
}

func getCommitParentRow(ctx context.Context, tx *pachsql.Tx, childCommit CommitID) (*Commit, error) {
	row := &Commit{}
	if err := tx.GetContext(ctx, row, fmt.Sprintf("%s WHERE ancestry.child=$1", getParentCommit), childCommit); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrParentCommitNotFound{ChildRowID: childCommit}
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
			return nil, ErrChildCommitNotFound{ParentRowID: parentCommit}
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
		return nil, ErrChildCommitNotFound{ParentRowID: parentCommit}
	}
	return children, nil
}

func getCommitRelativeRows(ctx context.Context, tx *pachsql.Tx, commitID CommitID) (*Commit, []*Commit, error) {
	commitParentRows, err := getCommitParentRow(ctx, tx, commitID)
	if err != nil && !errors.Is(ErrParentCommitNotFound{ChildRowID: commitID}, errors.Cause(err)) {
		return nil, nil, errors.Wrap(err, "getting parent commit")
		// if parent is missing, assume commit is root of a repo.
	}
	commitChildrenRows, err := getCommitChildrenRows(ctx, tx, commitID)
	if err != nil && !errors.Is(ErrChildCommitNotFound{ParentRowID: commitID}, errors.Cause(err)) {
		return nil, nil, errors.Wrap(err, "getting children commits")
		// if children is missing, assume commit is HEAD of some branch.
	}
	return commitParentRows, commitChildrenRows, nil
}

func getCommitRowByCommitKey(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) (*Commit, error) {
	row := &Commit{}
	if commit == nil {
		return nil, ErrCommitMissingInfo{Field: "Commit"}
	}
	id := CommitKey(commit)
	err := tx.QueryRowxContext(ctx, fmt.Sprintf("%s WHERE commit_id=$1", getCommit), id).StructScan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCommitNotFound{CommitID: id}
		}
		return nil, errors.Wrap(err, "scanning commit row")
	}
	return row, nil
}

func parseCommitInfoFromRow(row *Commit) (*pfs.CommitInfo, error) {
	commit, err := parseCommitFromRow(row)
	if err != nil {
		return nil, errors.Wrap(err, "getting commit from row")
	}
	commitInfo := &pfs.CommitInfo{
		Commit:      commit,
		Origin:      &pfs.CommitOrigin{Kind: pfs.OriginKind(pfs.OriginKind_value[strings.ToUpper(row.Origin)])},
		Started:     timeToTimestamppb(row.StartTime),
		Finishing:   timeToTimestamppb(row.FinishingTime),
		Finished:    timeToTimestamppb(row.FinishedTime),
		Description: row.Description,
		Details: &pfs.CommitInfo_Details{
			CompactingTime: bigIntToDurationpb(row.CompactingTime),
			ValidatingTime: bigIntToDurationpb(row.ValidatingTime),
			SizeBytes:      row.Size,
		},
	}
	return commitInfo, nil
}

func parseCommitFromRow(row *Commit) (*pfs.Commit, error) {
	repo := &pfs.Repo{
		Name: row.Repo.Name,
		Type: row.Repo.Type,
		Project: &pfs.Project{
			Name: row.Repo.Project.Name,
		},
	}
	parsedId := strings.Split(row.CommitID, "@")
	if len(parsedId) != 2 {
		return nil, errors.New(fmt.Sprintf("got invalid commit id from postgres: %s", row.CommitID))
	}
	commit := &pfs.Commit{
		Repo: repo,
		Id:   parsedId[1],
		Branch: &pfs.Branch{
			Repo: repo,
			Name: row.BranchName.String,
		},
	}
	return commit, nil
}

// CommitPair is an (id, commitInfo) tuple returned by the commit iterator.
type CommitPair struct {
	ID         CommitID
	CommitInfo *pfs.CommitInfo
}

// this dropped global variable instantiation forces the compiler to check whether CommitIterator implements stream.Iterator.
var _ stream.Iterator[CommitPair] = &CommitIterator{}

// CommitIterator batches a page of Commit entries along with their parent and children. (id, entry) tuples can be retrieved using iter.Next().
type CommitIterator struct {
	commitInfos map[CommitID]*pfs.CommitInfo
	db          *pachsql.DB
	gottenInfos int
	rev         bool
	limit       int
	offset      int
	commits     []Commit
	index       int
	filter      CommitListFilter
}

// Next advances the iterator by one row. It returns a stream.EOS when there are no more entries.
// The iterator prefetches the parents and children of the buffered commits until it hits an internal
// capacity.
func (iter *CommitIterator) Next(ctx context.Context, dst *CommitPair) error {
	if dst == nil {
		return errors.Wrap(fmt.Errorf("commit is nil"), "get next commit")
	}
	var err error
	if iter.index >= len(iter.commits) {
		iter.gottenInfos = 0
		iter.index = 0
		iter.offset += iter.limit
		if err := dbutil.WithTx(ctx, iter.db, func(ctx context.Context, tx *pachsql.Tx) error {
			iter.commits, err = listCommitPage(ctx, tx, iter.limit, iter.offset, iter.filter, iter.rev)
			if err != nil {
				return errors.Wrap(err, "list commit page")
			}
			if len(iter.commits) == 0 {
				return stream.EOS()
			}
			return errors.Wrap(iter.getCommitInfosForPageUntilCapacity(ctx, tx),
				"getting commit infos for page in list commits")
		}); err != nil {
			return err
		}
	} else if iter.index >= iter.gottenInfos {
		if err := dbutil.WithTx(ctx, iter.db, func(ctx context.Context, tx *pachsql.Tx) error {
			return errors.Wrap(iter.getCommitInfosForPageUntilCapacity(ctx, tx),
				"getting commit infos for page in list commits")
		}); err != nil {
			return err
		}
	}
	id := iter.commits[iter.index].ID
	*dst = CommitPair{
		CommitInfo: iter.commitInfos[id],
		ID:         id,
	}
	iter.index++
	return nil
}

// CommitFields is used in the ListCommitFilter and defines specific field names for type safety.
// This should hopefully prevent a library user from misconfiguring the filter.
type CommitFields string

const (
	CommitSetIDs   = CommitFields("commit_set_id")
	CommitOrigins  = CommitFields("origin")
	CommitRepos    = CommitFields("repo_id")
	CommitBranches = CommitFields("branch_id")
	CommitProjects = CommitFields("project_id")
)

// CommitListFilter is a filter for listing commits. It ANDs together separate keys, but ORs together the key values:
// where commit.<key_1> IN (<key_1:value_1>, <key_2:value_2>, ...) AND commit.<key_2> IN (<key_2:value_1>,<key_2:value_2>,...)
type CommitListFilter map[CommitFields][]string

// ListCommit returns a CommitIterator that exposes a Next() function for retrieving *pfs.CommitInfo references.
// It manages transactions on behalf of its user under the hood.
func ListCommit(ctx context.Context, db *pachsql.DB, filter CommitListFilter, rev bool) (*CommitIterator, error) {
	var iter *CommitIterator
	if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
		limit := 100
		page, err := listCommitPage(ctx, tx, limit, 0, filter, rev)
		if err != nil {
			return errors.Wrap(err, "new commits iterator")
		}
		iter = &CommitIterator{
			rev:         rev,
			commits:     page,
			limit:       limit,
			filter:      filter,
			gottenInfos: 0,
			db:          db,
		}
		return iter.getCommitInfosForPageUntilCapacity(ctx, tx)
	}); err != nil {
		return nil, errors.Wrap(err, "list commits first page")
	}
	return iter, nil
}

func (iter *CommitIterator) getCommitInfosForPageUntilCapacity(ctx context.Context, tx *pachsql.Tx) error {
	capacity := 500
	commitCount := 0
	iter.commitInfos = make(map[CommitID]*pfs.CommitInfo)
	for ; iter.gottenInfos < len(iter.commits); iter.gottenInfos++ {
		row := iter.commits[iter.gottenInfos]
		commit, err := getCommitInfoFromCommitRow(ctx, tx, &row)
		if err != nil {
			return errors.Wrap(err, "getting commitInfo from commit row")
		}
		commitCount += 1 + len(commit.ChildCommits)
		if commit.ParentCommit != nil {
			commitCount++
		}
		iter.commitInfos[row.ID] = commit
		if commitCount >= capacity {
			iter.gottenInfos++
			return nil
		}
	}
	return nil
}

func listCommitPage(ctx context.Context, tx *pachsql.Tx, limit, offset int, filter CommitListFilter, rev bool) ([]Commit, error) {
	var page []Commit
	where := ""
	conditions := make([]string, 0)
	query := getCommit
	order := "ASC"
	if rev {
		order = "DESC"
	}
	for key, vals := range filter {
		if len(vals) == 0 {
			continue
		}
		quotedVals := make([]string, 0)
		for _, val := range vals {
			quotedVals = append(quotedVals, fmt.Sprintf("'%s'", val))
		}
		switch key {
		case CommitBranches:
			conditions = append(conditions, fmt.Sprintf("commit.%s IN (SELECT id FROM pfs.branches WHERE name IN (%s))", string(key), strings.Join(quotedVals, ",")))
		case CommitRepos:
			conditions = append(conditions, fmt.Sprintf("commit.%s IN (SELECT id FROM pfs.repos WHERE name IN (%s))", string(key), strings.Join(quotedVals, ",")))
		case CommitProjects:
			conditions = append(conditions, fmt.Sprintf("repo.%s IN (SELECT id FROM core.projects WHERE name IN (%s))", string(key), strings.Join(quotedVals, ",")))
		default:
			conditions = append(conditions, fmt.Sprintf("commit.%s IN (%s)", string(key), strings.Join(quotedVals, ",")))
		}
	}
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}
	if err := tx.SelectContext(ctx, &page,
		fmt.Sprintf("%s %s GROUP BY commit.int_id, repo.name, repo.type, branch.name, project.name "+
			"ORDER BY commit.int_id %s LIMIT $1 OFFSET $2;", query, where, order), limit, offset); err != nil {
		return nil, errors.Wrap(err, "could not get commit page")
	}
	return page, nil
}

func sanitizeTimestamppb(timestamp *timestamppb.Timestamp) time.Time {
	if timestamp != nil {
		return timestamp.AsTime()
	}
	return time.Time{}
}

func durationpbToBigInt(duration *durationpb.Duration) int64 {
	if duration != nil {
		return duration.Seconds
	}
	return 0
}

func timeToTimestamppb(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func bigIntToDurationpb(s int64) *durationpb.Duration {
	if s == 0 {
		return nil
	}
	return durationpb.New(time.Duration(s))
}
