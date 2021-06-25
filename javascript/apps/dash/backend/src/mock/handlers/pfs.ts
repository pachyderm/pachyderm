import {Status} from '@grpc/grpc-js/build/src/constants';
import {Permission} from '@pachyderm/proto/pb/auth/auth_pb';
import {IAPIServer} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {RepoAuthInfo, RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {REPO_READER_PERMISSIONS} from '@dash-backend/constants/permissions';
import commits from '@dash-backend/mock/fixtures/commits';
import files from '@dash-backend/mock/fixtures/files';
import repos from '@dash-backend/mock/fixtures/repos';
import {createServiceError} from '@dash-backend/testHelpers';

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

const pfs: Pick<
  IAPIServer,
  'listRepo' | 'inspectRepo' | 'listCommit' | 'listFile'
> = {
  listRepo: (call) => {
    const [projectId] = call.metadata.get('project-id');
    const [accountId] = call.metadata.get('authn-token');
    const projectRepos = setAuthInfoForRepos(
      projectId ? repos[projectId.toString()] : repos['1'],
      accountId.toString(),
    );

    projectRepos.forEach((repo) => {
      call.write(repo);
    });

    call.end();
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
    const [accountId] = call.metadata.get('authn-token');

    const repoName = call.request.getRepo()?.getName();

    if (repoName && accountId) {
      const authInfo = repoAuthInfos[accountId.toString()][repoName];

      if (
        !authInfo.getPermissionsList().includes(Permission.REPO_LIST_COMMIT)
      ) {
        call.emit(
          'error',
          createServiceError({code: Status.UNKNOWN, details: 'not authorized'}),
        );
      }
    }

    const allCommits = commits[projectId.toString()] || commits['1'];

    allCommits.forEach((commit) => {
      if (commit.getCommit()?.getBranch()?.getRepo()?.getName() === repoName) {
        call.write(commit);
      }
    });

    call.end();
  },
  listFile: (call) => {
    const [projectId] = call.metadata.get('project-id');
    const replyFiles = projectId ? files[projectId.toString()] : files['1'];

    replyFiles.forEach((file) => call.write(file));
    call.end();
  },
};

export default pfs;
