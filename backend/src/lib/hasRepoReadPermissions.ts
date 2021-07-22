import {Permission} from '@pachyderm/proto/pb/auth/auth_pb';
import difference from 'lodash/difference';

import {REPO_READER_PERMISSIONS} from '@dash-backend/constants/permissions';

const hasRepoReadPermissions = (repoPermissions: Permission[] = []) => {
  return difference(REPO_READER_PERMISSIONS, repoPermissions).length === 0;
};

export default hasRepoReadPermissions;
