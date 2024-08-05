package server

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/pfs"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

func (a *apiServer) hasProjectAccess(
	ctx context.Context,
	txnCtx *txncontext.TransactionContext,
	repoInfo *pfs.RepoInfo,
	checkProjectAccess func(string) error) (bool, error) {
	if err := checkProjectAccess(repoInfo.Repo.Project.GetName()); err != nil {
		if !errors.As(err, &auth.ErrNotAuthorized{}) {
			return false, err
		}
		// Allow access if user has the right permissions at the individual Repo-level.
		if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, repoInfo.Repo, auth.Permission_REPO_READ); err != nil {
			if !errors.As(err, &auth.ErrNotAuthorized{}) {
				return false, errors.Wrapf(err, "could not check user is authorized to list repo, problem with repo %s", repoInfo.Repo)
			}
			// User does not have permissions, so we should skip this repo.
			return false, nil
		}
	}
	return true, nil
}

func (a *apiServer) createProject(ctx context.Context, req *pfs.CreateProjectRequest) error {
	return a.txnEnv.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		return a.createProjectInTransaction(ctx, txnCtx, req)
	})
}

func (a *apiServer) createProjectInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, req *pfs.CreateProjectRequest) error {
	if err := req.Project.ValidateName(); err != nil {
		return errors.Wrapf(err, "invalid project name")
	}
	if req.Update {
		src, err := pfsdb.GetProjectInfoByName(ctx, txnCtx.SqlTx, req.GetProject().GetName())
		if err != nil {
			return errors.Wrapf(err, "get project %v", req.GetProject().GetName())
		}
		dst := &pfs.ProjectInfo{
			Project:     req.Project,
			Description: req.Description,
			Metadata:    src.Metadata,
		}
		return errors.Wrapf(
			pfsdb.UpsertProject(ctx, txnCtx.SqlTx, dst),
			"update project %s", req.GetProject().GetName())
	}
	// If auth is active, make caller the owner of this new project.
	var username string
	if whoAmI, err := txnCtx.WhoAmI(); err == nil {
		username = whoAmI.GetUsername()
		if err := a.env.Auth.CreateRoleBindingInTransaction(
			ctx,
			txnCtx,
			whoAmI.Username,
			[]string{auth.ProjectOwnerRole},
			&auth.Resource{Type: auth.ResourceType_PROJECT, Name: req.Project.GetName()},
		); err != nil && !errors.Is(err, col.ErrExists{}) {
			return errors.Wrapf(err, "could not create role binding for new project %s", req.Project.GetName())
		}
	} else if !errors.Is(err, auth.ErrNotActivated) {
		return errors.Wrap(err, "could not get caller's username")
	}
	if err := pfsdb.CreateProject(ctx, txnCtx.SqlTx, &pfs.ProjectInfo{
		Project:     req.Project,
		Description: req.Description,
		CreatedAt:   timestamppb.Now(),
		CreatedBy:   username,
	}); err != nil {
		if errors.As(err, &pfsdb.ProjectAlreadyExistsError{}) {
			return errors.Join(err, pfsserver.ErrProjectExists{Project: req.Project})
		}
		return errors.Wrap(err, "could not create project")
	}
	return nil
}

func (a *apiServer) inspectProject(ctx context.Context, project *pfs.Project) (*pfs.ProjectInfo, error) {
	var pi *pfs.ProjectInfo
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		pi, err = pfsdb.GetProjectInfoByName(ctx, tx, pfsdb.ProjectKey(project))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, errors.Wrapf(err, "error getting project %q", project)
	}
	resp, err := a.env.Auth.GetPermissions(ctx, &auth.GetPermissionsRequest{Resource: project.AuthResource()})
	if err != nil {
		if errors.Is(err, auth.ErrNotActivated) {
			return pi, nil
		}
		return nil, errors.Wrapf(err, "error getting permissions for project %q", project)
	}
	pi.AuthInfo = &pfs.AuthInfo{Permissions: resp.Permissions, Roles: resp.Roles}
	return pi, nil
}

// The ProjectInfo provided to the closure is repurposed on each invocation, so it's the client's responsibility to clone the ProjectInfo if desired
func (a *apiServer) listProject(ctx context.Context, cb func(*pfs.ProjectInfo) error) error {
	authIsActive := true
	return errors.Wrap(a.txnEnv.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		return a.listProjectInTransaction(ctx, txnCtx, func(proj *pfs.ProjectInfo) error {
			if authIsActive {
				resp, err := a.env.Auth.GetPermissionsInTransaction(ctx, txnCtx, &auth.GetPermissionsRequest{Resource: proj.GetProject().AuthResource()})
				if err != nil {
					if errors.Is(err, auth.ErrNotActivated) {
						// Avoid unnecessary subsequent Auth API calls.
						authIsActive = false
						return cb(proj)
					}
					return errors.Wrapf(err, "getting permissions for project %s", proj.Project)
				}
				proj.AuthInfo = &pfs.AuthInfo{Permissions: resp.Permissions, Roles: resp.Roles}
			}
			return cb(proj)
		})
	}), "list projects")
}

// The ProjectInfo provided to the closure is repurposed on each invocation, so it's the client's responsibility to clone the ProjectInfo if desired
func (a *apiServer) listProjectInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, cb func(*pfs.ProjectInfo) error) error {
	return errors.Wrap(pfsdb.ForEachProject(ctx, txnCtx.SqlTx, func(project pfsdb.Project) error {
		return cb(project.ProjectInfo)
	}), "list projects in transaction")
}

// TODO: delete all repos and pipelines within project
func (a *apiServer) deleteProject(ctx context.Context, txnCtx *txncontext.TransactionContext, project *pfs.Project, force bool) error {
	if err := project.ValidateName(); err != nil {
		return errors.Wrap(err, "invalid project name")
	}
	if err := a.env.Auth.CheckProjectIsAuthorizedInTransaction(ctx, txnCtx, project, auth.Permission_PROJECT_DELETE, auth.Permission_PROJECT_MODIFY_BINDINGS); err != nil {
		return errors.Wrapf(err, "user is not authorized to delete project %q", project)
	}
	var errs error
	repos, err := a.listRepoInTransaction(ctx, txnCtx, false, "", []*pfs.Project{project}, nil)
	if err != nil {
		return errors.Wrap(err, "list repos to determine if any still exist")
	}
	for _, repoInfo := range repos {
		errs = errors.Join(errs, fmt.Errorf("repo %v still exists", repoInfo.GetRepo()))
	}
	if errs != nil && !force {
		return status.Error(codes.FailedPrecondition, fmt.Sprintf("cannot delete project %s: %v", project.Name, errs))
	}
	if err := pfsdb.DeleteProject(ctx, txnCtx.SqlTx, pfsdb.ProjectKey(project)); err != nil {
		return errors.Wrapf(err, "delete project %q", project)
	}
	if err := a.env.Auth.DeleteRoleBindingInTransaction(ctx, txnCtx, project.AuthResource()); err != nil {
		if !errors.Is(err, auth.ErrNotActivated) {
			return errors.Wrapf(err, "delete role binding for project %q", project)
		}
	}
	return nil
}

func (a *apiServer) deleteAll(ctx context.Context) error {
	return a.txnEnv.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		if _, err := a.deleteReposInTransaction(ctx, txnCtx, nil /* projects */, true /* force */); err != nil {
			return errors.Wrap(err, "could not delete all repos")
		}
		if err := a.listProjectInTransaction(ctx, txnCtx, func(pi *pfs.ProjectInfo) error {
			return errors.Wrapf(a.deleteProject(ctx, txnCtx, pi.Project, true /* force */), "delete project %q", pi.Project.String())
		}); err != nil {
			return err
		} // now that the cluster is empty, recreate the default project
		return a.createProjectInTransaction(ctx, txnCtx, &pfs.CreateProjectRequest{Project: &pfs.Project{Name: "default"}})
	})
}
