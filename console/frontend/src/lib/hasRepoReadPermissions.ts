import difference from 'lodash/difference';

import {Permission} from '@dash-frontend/api/auth';
import {REPO_READER_PERMISSIONS} from '@dash-frontend/constants/permissions';

const hasRepoReadPermissions = (repoPermissions?: Permission[]) => {
  if (!repoPermissions) return true;
  else return difference(REPO_READER_PERMISSIONS, repoPermissions).length === 0;
};

export default hasRepoReadPermissions;
