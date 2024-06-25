import {RepoInfo} from '@dash-frontend/api/pfs';
import {useBranches} from '@dash-frontend/hooks/useBranches';
import {useCommits} from '@dash-frontend/hooks/useCommits';
import useFileBrowserNavigation from '@dash-frontend/hooks/useFileBrowserNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';

export const COMMIT_LIMIT = 6;

const useCommitList = (repo?: RepoInfo) => {
  const {projectId, repoId} = useUrlState();
  const {getPathToFileBrowser} = useFileBrowserNavigation();

  const {branches} = useBranches({
    projectId,
    repoId,
  });
  const {commits, loading, error} = useCommits(
    {
      projectName: projectId,
      repoName: repoId,
      args: {
        number: COMMIT_LIMIT,
      },
    },
    !repo,
  );
  const [, ...previousCommits] = commits || [];

  return {
    loading,
    error,
    commits,
    branches,
    previousCommits,
    getPathToFileBrowser,
    projectId,
    repoId,
  };
};

export default useCommitList;
