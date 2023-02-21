import {ReposQueryArgs} from '@graphqlTypes';

import {REPO_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useReposQuery} from '@dash-frontend/generated/hooks';

const useRepos = (args: ReposQueryArgs) => {
  const {data, error, loading} = useReposQuery({
    variables: {args},
    pollInterval: REPO_POLL_INTERVAL_MS,
  });

  return {
    repos: data?.repos,
    error,
    loading,
  };
};

export default useRepos;
