import {RepoAuthInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {REPO_READER_PERMISSIONS} from '@dash-backend/constants/permissions';

import allRepos from './repos';

export interface RepoAuthInfoMap {
  [repoId: string]: RepoAuthInfo;
}

const repoReaderAuthInfo = Object.values(allRepos).reduce<{
  [repoId: string]: RepoAuthInfo;
}>((acc, projectRepos) => {
  projectRepos.forEach((repo) => {
    acc[repo.getRepo()?.getName() || ''] =
      new RepoAuthInfo().setPermissionsList(REPO_READER_PERMISSIONS);
  });

  return acc;
}, {});

const repoAuthInfos: {[accountId: string]: RepoAuthInfoMap} = {
  default: repoReaderAuthInfo,
  '1': repoReaderAuthInfo,
  '2': {
    ...repoReaderAuthInfo,
    montage: new RepoAuthInfo().setPermissionsList([]),
  },
};

export default repoAuthInfos;
