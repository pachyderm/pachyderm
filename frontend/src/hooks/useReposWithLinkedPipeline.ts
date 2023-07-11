import {ReposQueryArgs} from '@graphqlTypes';

import {REPO_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useReposWithLinkedPipelineQuery} from '@dash-frontend/generated/hooks';

const useReposWithLinkedPipeline = (args: ReposQueryArgs) => {
  const {data, error, loading} = useReposWithLinkedPipelineQuery({
    variables: {args},
    pollInterval: REPO_POLL_INTERVAL_MS,
  });

  return {
    repos: data?.repos,
    error,
    loading,
  };
};

export default useReposWithLinkedPipeline;
