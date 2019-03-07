package server

import (
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

type Operation interface {
    // TODO: probably don't need to pass the full client here
    validate(pachClient *client.APIClient) error
    execute(pachClient *client.APIClient, stm col.STM) error
}

type CreateBranchOp struct {
    branch *pfs.Branch
    commit *pfs.Commit
    provenance []*pfs.Branch
}

func (op CreateBranchOp) validate(pachClient *client.APIClient) error {
    return nil
}

func (op CreateBranchOp) execute(pachClient *client.APIClient, stm col.STM) error {
    return nil
}
