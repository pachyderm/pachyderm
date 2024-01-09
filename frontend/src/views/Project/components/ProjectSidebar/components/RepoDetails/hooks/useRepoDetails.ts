import {useMemo} from 'react';

import {
  Permission,
  ResourceType,
  useAuthorize,
} from '@dash-frontend/hooks/useAuthorize';
import {useCommitDiff} from '@dash-frontend/hooks/useCommitDiff';
import {useCommits} from '@dash-frontend/hooks/useCommits';
import {useCurrentRepo} from '@dash-frontend/hooks/useCurrentRepo';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import formatDiff from '@dash-frontend/lib/formatDiff';

const useRepoDetails = () => {
  const {repoId, projectId, commitId} = useUrlState();
  const {searchParams} = useUrlQueryState();
  const {loading: repoLoading, repo, error: repoError} = useCurrentRepo();
  const {getPathToFileBrowser} = useFileBrowserNavigation();

  const {
    commits,
    loading: commitLoading,
    error: commitError,
  } = useCommits({
    projectName: projectId,
    repoName: repoId,
    args: {
      commitIdCursor: searchParams.globalIdFilter || commitId || '',
      number: 1,
    },
  });

  const commit = commits && commits[0];

  const {
    fileDiff,
    loading: diffLoading,
    error: diffError,
  } = useCommitDiff(
    {
      newFile: {
        commit: {
          id: commit?.commit?.id,
          repo: {
            name: commit?.commit?.repo?.name,
            project: {
              name: projectId,
            },
            type: 'user',
          },
          branch: {
            name: commit?.commit?.branch?.name,
          },
        },
        path: '/',
      },
    },
    Boolean(commit && commit.finished),
  );

  const formattedDiff = useMemo(() => formatDiff(fileDiff || []), [fileDiff]);

  const {
    hasRepoModifyBindings: hasRepoEditRoles,
    hasRepoRead,
    hasRepoWrite,
  } = useAuthorize({
    permissions: [
      Permission.REPO_MODIFY_BINDINGS,
      Permission.REPO_READ,
      Permission.REPO_WRITE,
    ],
    resource: {type: ResourceType.REPO, name: `${projectId}/${repoId}`},
  });

  const currentRepoLoading =
    repoLoading || commitLoading || repoId !== repo?.repo?.name;

  return {
    projectId,
    repoId,
    repo,
    commit,
    currentRepoLoading,
    commitError,
    commitDiff: formattedDiff,
    diffLoading,
    diffError,
    repoError,
    hasRepoEditRoles,
    getPathToFileBrowser,
    hasRepoRead,
    hasRepoWrite,
    globalId: searchParams.globalIdFilter,
  };
};

export default useRepoDetails;
