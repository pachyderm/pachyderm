import {Permission} from '@pachyderm/node-pachyderm';
import difference from 'lodash/difference';

import {REPO_READER_PERMISSIONS} from '@dash-backend/constants/permissions';

const hasRepoReadPermissions = (repoPermissions: Permission[] = []) => {
  return difference(REPO_READER_PERMISSIONS, repoPermissions).length === 0;
};

export default hasRepoReadPermissions;
