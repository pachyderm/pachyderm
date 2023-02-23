import {RepoQuery, OriginKind} from '@graphqlTypes';
import {useCallback} from 'react';

import useCommits from '@dash-frontend/hooks/useCommits';
import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';

export const COMMIT_LIMIT = 6;

const useCommitList = (repo?: RepoQuery['repo']) => {
  const {projectId, repoId} = useUrlState();
  const [hideAutoCommits, handleHideAutoCommitChange] = useLocalProjectSettings(
    {projectId, key: 'hide_auto_commits'},
  );

  const handleOriginFilter = useCallback(
    (selection: string) => {
      if (selection !== 'all-origin') {
        handleHideAutoCommitChange(true);
      } else {
        handleHideAutoCommitChange(false);
      }
    },
    [handleHideAutoCommitChange],
  );

  const {commits, loading, error} = useCommits({
    args: {
      projectId,
      repoName: repoId,
      pipelineName: repo?.linkedPipeline?.name,
      originKind: hideAutoCommits ? OriginKind.USER : undefined,
      number: COMMIT_LIMIT,
    },
    skip: !repo || repo.branches.length === 0,
  });
  const [, ...previousCommits] = commits || [];

  return {
    loading,
    error,
    commits,
    handleOriginFilter,
    previousCommits,
    hideAutoCommits,
  };
};

export default useCommitList;
