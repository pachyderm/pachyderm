package clientsdk

import (
	"errors"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func ForEachRepoInfo(client pfs.API_ListRepoClient, cb func(*pfs.RepoInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := cb(x); err != nil {
			if err == pacherr.ErrBreak {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ListRepoInfo(client pfs.API_ListRepoClient) ([]*pfs.RepoInfo, error) {
	var results []*pfs.RepoInfo
	if err := ForEachRepoInfo(client, func(x *pfs.RepoInfo) error {
		results = append(results, x)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func ForEachBranchInfo(client pfs.API_ListBranchClient, cb func(*pfs.BranchInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := cb(x); err != nil {
			if err == pacherr.ErrBreak {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ListBranchInfo(client pfs.API_ListBranchClient) ([]*pfs.BranchInfo, error) {
	var results []*pfs.BranchInfo
	if err := ForEachBranchInfo(client, func(x *pfs.BranchInfo) error {
		results = append(results, x)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func ForEachCommitInfo(client pfs.API_ListCommitClient, cb func(*pfs.CommitInfo) error) error {
	for {
		ci, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := cb(ci); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				return nil
			}
			return err
		}
	}
	return nil
}

func ListCommitInfo(client pfs.API_ListCommitClient) ([]*pfs.CommitInfo, error) {
	var results []*pfs.CommitInfo
	if err := ForEachCommitInfo(client, func(ci *pfs.CommitInfo) error {
		results = append(results, ci)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}
