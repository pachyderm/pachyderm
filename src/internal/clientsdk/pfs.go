package clientsdk

import (
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func ForEachBranchInfo(client pfs.API_ListBranchClient, cb func(*pfs.BranchInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
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

func ForEachRepoInfo(client pfs.API_ListRepoClient, cb func(*pfs.RepoInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
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

func ForEachProjectInfo(client pfs.API_ListProjectClient, cb func(*pfs.ProjectInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ListProjectInfo(client pfs.API_ListProjectClient) ([]*pfs.ProjectInfo, error) {
	var results []*pfs.ProjectInfo
	if err := ForEachProjectInfo(client, func(x *pfs.ProjectInfo) error {
		results = append(results, x)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func ForEachCommitSet(client pfs.API_ListCommitSetClient, cb func(*pfs.CommitSetInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ForEachCommit(client pfs.API_ListCommitClient, cb func(*pfs.CommitInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ListCommit(client pfs.API_ListCommitClient) ([]*pfs.CommitInfo, error) {
	var results []*pfs.CommitInfo
	if err := ForEachCommit(client, func(x *pfs.CommitInfo) error {
		results = append(results, x)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func ForEachFile(client pfs.API_ListFileClient, cb func(*pfs.FileInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ListFile(client pfs.API_ListFileClient) ([]*pfs.FileInfo, error) {
	var results []*pfs.FileInfo
	if err := ForEachFile(client, func(x *pfs.FileInfo) error {
		results = append(results, x)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func ForEachSubscribeCommit(client pfs.API_SubscribeCommitClient, cb func(*pfs.CommitInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}

func ForEachGlobFile(client pfs.API_GlobFileClient, cb func(*pfs.FileInfo) error) error {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.EnsureStack(err)
		}
		if err := cb(x); err != nil {
			if errors.Is(err, pacherr.ErrBreak) {
				err = nil
			}
			return err
		}
	}
	return nil
}
