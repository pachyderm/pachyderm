import {Status} from '@grpc/grpc-js/build/src/constants';
import {IAPIServer} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {
  ListRepoResponse,
  RepoAuthInfo,
  RepoInfo,
} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {REPO_READER_PERMISSIONS} from '@dash-backend/constants/permissions';
import commits from '@dash-backend/mock/fixtures/commits';
import repos from '@dash-backend/mock/fixtures/repos';

import repoAuthInfos from '../fixtures/repoAuthInfos';

const setAuthInfoForRepo = (repo: RepoInfo, accountId = '') => {
  const repoName = repo.getRepo()?.getName();
  const authInfo = repoAuthInfos[accountId] || repoAuthInfos['default'];

  repo.setAuthInfo(
    repoName
      ? authInfo[repoName]
      : new RepoAuthInfo().setPermissionsList(REPO_READER_PERMISSIONS),
  );

  return repo;
};

const setAuthInfoForRepos = (repos: RepoInfo[], accountId = '') => {
  repos.forEach((repo) => {
    setAuthInfoForRepo(repo, accountId);
  });

  return repos;
};

const pfs: Pick<IAPIServer, 'listRepo' | 'inspectRepo' | 'listCommit'> = {
  listRepo: (call, callback) => {
    const [projectId] = call.metadata.get('project-id');
    const [accountId] = call.metadata.get('authn-token');

    const reply = new ListRepoResponse();

    const projectRepos = setAuthInfoForRepos(
      projectId ? repos[projectId.toString()] : repos['1'],
      accountId.toString(),
    );

    reply.setRepoInfoList(projectRepos);

    callback(null, reply);
  },
  inspectRepo: (call, callback) => {
    const [projectId] = call.metadata.get('project-id');
    const [accountId] = call.metadata.get('authn-token');

    const repoName = call.request.getRepo()?.getName();
    const repo = (
      projectId ? repos[projectId.toString()] : repos['tutorial']
    ).find((r) => r.getRepo()?.getName() === repoName);

    if (repo) {
      setAuthInfoForRepo(repo, accountId.toString());
      callback(null, repo);
    } else {
      callback({code: Status.NOT_FOUND, details: 'repo not found'});
    }
  },
  listCommit: (call) => {
    const [projectId] = call.metadata.get('project-id');
    const repoName = call.request.getRepo()?.getName();
    const allCommits = commits[projectId.toString()] || commits['1'];

    allCommits.forEach((commit) => {
      if (commit.getCommit()?.getBranch()?.getRepo()?.getName() === repoName) {
        call.write(commit);
      }
    });

    call.end();
  },
};

export default pfs;
