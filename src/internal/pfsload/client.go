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
	WaitCommitSet(ctx context.Context, id string, cb func(*pfs.CommitInfo) error) error
}

type pachClient struct {
	pfs pfs.APIClient
}

func NewPachClient(pfs pfs.APIClient) Client {
	return &pachClient{pfs: pfs}
}

func (pc *pachClient) WithCreateFileSetClient(ctx context.Context, cb func(client.ModifyFile) error) (*pfs.CreateFileSetResponse, error) {
	return client.WithCreateFileSetClient(ctx, pc.pfs, cb)
}

func (pc *pachClient) AddFileSet(ctx context.Context, commit *pfs.Commit, filesetID string) error {
	project := commit.Repo.Project.GetName()
	repo := commit.Repo.Name
	branch := commit.Branch.Name
	return client.AddFileSet(ctx, pc.pfs, project, repo, branch, commit.Id, filesetID)
}

func (pc *pachClient) GlobFile(ctx context.Context, commit *pfs.Commit, pattern string, cb func(*pfs.FileInfo) error) error {
	return client.GlobFile(ctx, pc.pfs, commit, pattern, cb)
}

func (pc *pachClient) WaitCommitSet(ctx context.Context, id string, cb func(*pfs.CommitInfo) error) error {
	return client.WaitCommitSet(ctx, pc.pfs, id, cb)
}
