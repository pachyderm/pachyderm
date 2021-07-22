import {Permission} from '@pachyderm/proto/pb/auth/auth_pb';

export const REPO_READER_PERMISSIONS = [
  Permission.REPO_READ,
  Permission.REPO_INSPECT_COMMIT,
  Permission.REPO_LIST_COMMIT,
  Permission.REPO_LIST_BRANCH,
  Permission.REPO_LIST_FILE,
  Permission.REPO_INSPECT_FILE,
  Permission.REPO_ADD_PIPELINE_READER,
  Permission.REPO_REMOVE_PIPELINE_READER,
  Permission.PIPELINE_LIST_JOB,
];
