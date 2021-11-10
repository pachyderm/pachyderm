import {CommitInfo} from '@pachyderm/node-pachyderm';

import formatBytes from '@dash-backend/lib/formatBytes';
import getSizeBytes from '@dash-backend/lib/getSizeBytes';
import {toGQLCommitOrigin} from '@dash-backend/lib/gqlEnumMappers';
import {Commit} from '@graphqlTypes';

export const commitInfoToGQLCommit = (commit: CommitInfo.AsObject): Commit => {
  return {
    repoName: commit.commit?.branch?.repo?.name || '',
    branch: commit.commit?.branch
      ? {
          name: commit.commit?.branch?.name || '',
        }
      : undefined,
    originKind: toGQLCommitOrigin(commit.origin?.kind),
    description: commit.description,
    finished: commit.finished?.seconds || -1,
    id: commit.commit?.id || '',
    started: commit.started?.seconds || -1,
    sizeBytes: getSizeBytes(commit),
    sizeDisplay: formatBytes(getSizeBytes(commit)),
    hasLinkedJob: false,
  };
};
