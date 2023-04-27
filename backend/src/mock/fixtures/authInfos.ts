import {REPO_READER_PERMISSIONS} from '@dash-backend/constants/permissions';
import {AuthInfo} from '@dash-backend/proto';

import allRepos from './repos';

export interface RepoAuthInfoMap {
  [repoId: string]: AuthInfo;
}

const repoReaderAuthInfo = Object.values(allRepos).reduce<{
  [repoId: string]: AuthInfo;
}>((acc, projectRepos) => {
  projectRepos.forEach((repo) => {
    acc[repo.getRepo()?.getName() || ''] = new AuthInfo().setPermissionsList(
      REPO_READER_PERMISSIONS,
    );
  });

  return acc;
}, {});

const repoAuthInfos: {[accountId: string]: RepoAuthInfoMap} = {
  default: repoReaderAuthInfo,
  '1': repoReaderAuthInfo,
  '2': {
    ...repoReaderAuthInfo,
    montage: new AuthInfo().setPermissionsList([]),
  },
};

export default repoAuthInfos;
