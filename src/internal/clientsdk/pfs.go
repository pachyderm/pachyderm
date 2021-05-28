package clientsdk

import (
	"io"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func ForEachBranchInfo(client pfs.API_ListBranchClient, cb func(*pfs.BranchInfo) error) (retErr error) {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := cb(x); err != nil {
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

func ForEachRepoInfo(client pfs.API_ListRepoClient, cb func(*pfs.RepoInfo) error) (retErr error) {
	for {
		x, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := cb(x); err != nil {
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
