import {RepoQueryArgs} from '@graphqlTypes';

import {REPO_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useRepoWithCommitQuery} from '@dash-frontend/generated/hooks';

const useRepoWithCommit = (args: RepoQueryArgs) => {
  const repoQuery = useRepoWithCommitQuery({
    variables: {args},
    pollInterval: REPO_POLL_INTERVAL_MS,
  });

  return {
    ...repoQuery,
    repos: repoQuery.data?.repo,
  };
};

export default useRepoWithCommit;
