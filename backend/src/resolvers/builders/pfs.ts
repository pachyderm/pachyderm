import formatBytes from '@dash-backend/lib/formatBytes';
import getSizeBytes from '@dash-backend/lib/getSizeBytes';
import {toGQLCommitOrigin} from '@dash-backend/lib/gqlEnumMappers';
import {Commit, CommitInfo} from '@dash-backend/proto';
import {Commit as graphqlCommit} from '@graphqlTypes';

export const commitInfoToGQLCommit = (
  commit: CommitInfo.AsObject,
): graphqlCommit => {
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

export const commitToGQLCommit = (commit: Commit.AsObject) => {
  const branch = {name: '', repo: {name: '', type: ''}};

  if (commit.branch) {
    branch.name = commit.branch.name;
    if (commit.branch.repo) {
      branch.repo.name = commit.branch.repo.name;
      branch.repo.type = commit.branch.repo.type;
    }
  }
  return {
    id: commit.id,
    branch,
  };
};
