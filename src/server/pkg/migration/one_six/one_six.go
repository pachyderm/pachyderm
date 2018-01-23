package one_six

import (
	"context"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/pfsdb"
)

type Client struct {
	etcdClient *etcd.Client
	repos      col.Collection
	commits    func(string) col.Collection
}

func NewClient(etcdClient *etcd.Client, etcdPrefix string) *Client {
	return &Client{
		etcdClient: etcdClient,
		repos:      pfsdb.Repos(etcdClient, etcdPrefix),
		commits:    func(repo string) col.Collection { return pfsdb.Commits(etcdClient, etcdPrefix, repo) },
	}
}

type RepoInfo = pfs.RepoInfo
type CommitInfo = pfs.CommitInfo

// Migrate extracts pachyderm structures and calls the callback function with
// each one.
func (c *Client) Migrate(ctx context.Context, repoF func(*RepoInfo, col.STM) error, commitF func(*CommitInfo, col.STM) error) error {
	rIt, err := c.repos.ReadOnly(ctx).List()
	if err != nil {
		return err
	}
	for {
		repoName, repoInfo := "", &RepoInfo{}
		ok, err := rIt.Next(&repoName, repoInfo)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		if _, err := col.NewSTM(ctx, c.etcdClient, func(stm col.STM) error {
			if err := c.repos.ReadWrite(stm).Delete(repoName); err != nil {
				return err
			}
			return repoF(repoInfo, stm)
		}); err != nil {
			return err
		}
		commits := c.commits(repoName)
		cIt, err := commits.ReadOnly(ctx).List()
		if err != nil {
			return err
		}
		for {
			commitID, commitInfo := "", &CommitInfo{}
			ok, err := cIt.Next(&commitID, commitInfo)
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			if _, err := col.NewSTM(ctx, c.etcdClient, func(stm col.STM) error {
				if err := commits.ReadWrite(stm).Delete(commitID); err != nil {
					return err
				}
				return commitF(commitInfo, stm)
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
