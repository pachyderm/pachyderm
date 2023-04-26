import {RepoQuery} from '@graphqlTypes';

import useCommits from '@dash-frontend/hooks/useCommits';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';

export const COMMIT_LIMIT = 6;

const useCommitList = (repo?: RepoQuery['repo']) => {
  const {projectId, repoId} = useUrlState();
  const {getPathToFileBrowser} = useFileBrowserNavigation();

  const {commits, loading, error} = useCommits({
    args: {
      projectId,
      repoName: repoId,
      originKind: undefined,
      number: COMMIT_LIMIT,
    },
    skip: !repo || repo.branches.length === 0,
  });
  const [, ...previousCommits] = commits || [];

  return {
    loading,
    error,
    commits,
    previousCommits,
    getPathToFileBrowser,
    projectId,
    repoId,
  };
};

export default useCommitList;
