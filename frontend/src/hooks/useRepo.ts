import {REPO_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useRepoQuery} from '@dash-frontend/generated/hooks';
import {RepoQueryArgs} from '@graphqlTypes';

const useRepo = (args: RepoQueryArgs) => {
  const {data, error, loading} = useRepoQuery({
    variables: {args},
    pollInterval: REPO_POLL_INTERVAL_MS,
  });

  return {
    repo: data?.repo,
    error,
    loading,
  };
};

export default useRepo;
