package pfsload

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// Client is the standard interface for a load testing client.
// TODO: This should become the client.Client interface when we put the standard pach client behind an interface that
// takes a context as the first parameter for each method.
type Client interface {
	WithCreateFileSetClient(ctx context.Context, cb func(client.ModifyFile) error) (*pfs.CreateFileSetResponse, error)
	AddFileSet(ctx context.Context, commit *pfs.Commit, ID string) error
	GlobFile(ctx context.Context, commit *pfs.Commit, pattern string, cb func(*pfs.FileInfo) error) error
	WaitCommitSet(id string, cb func(*pfs.CommitInfo) error) error
	Ctx() context.Context
}

type pachClient struct {
	client *client.APIClient
}

func NewPachClient(client *client.APIClient) Client {
	return &pachClient{client: client}
}

func (pc *pachClient) WithCreateFileSetClient(ctx context.Context, cb func(client.ModifyFile) error) (*pfs.CreateFileSetResponse, error) {
	ctx = pc.client.AddMetadata(ctx)
	return pc.client.WithCtx(ctx).WithCreateFileSetClient(cb)
}

func (pc *pachClient) AddFileSet(ctx context.Context, commit *pfs.Commit, filesetID string) error {
	project := commit.Repo.Project.GetName()
	repo := commit.Repo.Name
	branch := commit.Branch.Name
	return pc.client.WithCtx(ctx).AddFileSet(project, repo, branch, commit.Id, filesetID)
}

func (pc *pachClient) GlobFile(ctx context.Context, commit *pfs.Commit, pattern string, cb func(*pfs.FileInfo) error) error {
	return pc.client.WithCtx(ctx).GlobFile(commit, pattern, cb)
}

func (pc *pachClient) WaitCommitSet(id string, cb func(*pfs.CommitInfo) error) error {
	return pc.client.WaitCommitSet(id, cb)
}

func (pc *pachClient) Ctx() context.Context {
	return pc.client.Ctx()
}
