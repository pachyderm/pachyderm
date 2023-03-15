import {ReposQueryArgs} from '@graphqlTypes';

import {REPO_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useReposWithCommitQuery} from '@dash-frontend/generated/hooks';

const useReposWithCommit = (args: ReposQueryArgs) => {
  const reposQuery = useReposWithCommitQuery({
    variables: {args},
    pollInterval: REPO_POLL_INTERVAL_MS,
  });

  return {
    ...reposQuery,
    repos: reposQuery.data?.repos,
  };
};

export default useReposWithCommit;
