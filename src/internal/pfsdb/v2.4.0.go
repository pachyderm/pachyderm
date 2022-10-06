package pfsdb

import (
	"context"
	"strings"

	"github.com/gogo/protobuf/proto"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// returns collections released in v2.4.0 - specifically the projects collection
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CollectionsV2_4_0() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(projectsCollectionName, nil, nil, nil, nil),
	}
}

// MigrateV2_4_0 migrates PFS to be fully project-aware with a default project.
// It uses some internal knowledge about how cols.PostgresCollection works to do
// so.
func MigrateV2_4_0(ctx context.Context, tx *pachsql.Tx) error {
	var oldRepo = new(pfs.Repo)
	if err := col.MigratePostgreSQLCollection(ctx, tx, "repos", reposIndexes, oldRepo, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		if oldRepo.Project.GetName() == "" {
			oldRepo.Project = &pfs.Project{Name: "default"}
		}
		return RepoKey(oldRepo), oldRepo, nil

	},
		col.WithKeyCheck(repoKeyCheck),
		col.WithKeyGen(func(key interface{}) (string, error) {
			if repo, ok := key.(*pfs.Repo); !ok {
				return "", errors.New("key must be a repo")
			} else {
				return RepoKey(repo), nil
			}
		}),
	); err != nil {
		return errors.Wrap(err, "could not migrate repos")
	}
	var oldBranch = new(pfs.Branch)
	if err := col.MigratePostgreSQLCollection(ctx, tx, "branches", branchesIndexes, oldBranch, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		if oldBranch.Repo.Project.GetName() == "" {
			oldBranch.Repo.Project = &pfs.Project{Name: "default"}
		}
		return BranchKey(oldBranch), oldBranch, nil

	},
		col.WithKeyGen(func(key interface{}) (string, error) {
			if branch, ok := key.(*pfs.Branch); !ok {
				return "", errors.New("key must be a branch")
			} else {
				return BranchKey(branch), nil
			}
		}),
		col.WithKeyCheck(func(key string) error {
			keyParts := strings.Split(key, "@")
			if len(keyParts) != 2 {
				return errors.Errorf("branch key %s isn't valid, use BranchKey to generate it", key)
			}
			if uuid.IsUUIDWithoutDashes(keyParts[1]) {
				return errors.Errorf("branch name cannot be a UUID V4")
			}
			return repoKeyCheck(keyParts[0])
		})); err != nil {
		return errors.Wrap(err, "could not migrate branches")
	}
	var oldCommit = new(pfs.Commit)
	if err := col.MigratePostgreSQLCollection(ctx, tx, "commits", commitsIndexes, oldCommit, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		if b := oldCommit.Branch; b != nil {
			if r := b.Repo; r != nil {
				if r.Project.GetName() == "" {
					r.Project = &pfs.Project{Name: "default"}
				}
			}
		}
		return CommitKey(oldCommit), oldCommit, nil

	}, col.WithKeyGen(func(key interface{}) (string, error) {
		if commit, ok := key.(*pfs.Commit); !ok {
			return "", errors.New("key must be a commit")
		} else {
			return CommitKey(commit), nil
		}
	})); err != nil {
		return errors.Wrap(err, "could not migrate commits")
	}
	var oldProject = new(pfs.Project)
	if err := col.MigratePostgreSQLCollection(ctx, tx, "projects", nil, oldProject, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		if oldProject.Name == "" {
			oldProject.Name = "default"
			return "default", oldProject, nil
		}
		return ProjectKey(oldProject), oldProject, nil
	}, col.WithKeyGen(func(key interface{}) (string, error) {
		if project, ok := key.(*pfs.Project); !ok {
			return "", errors.New("key must be a project")
		} else {
			return ProjectKey(project), nil
		}
	})); err != nil {
		return errors.Wrap(err, "could not migrate projects")
	}

	return nil
}
