import {RepoQueryArgs} from '@graphqlTypes';

import {REPO_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useRepoWithLinkedPipelineQuery} from '@dash-frontend/generated/hooks';

const useRepoWithLinkedPipeline = (args: RepoQueryArgs) => {
  const {data, error, loading} = useRepoWithLinkedPipelineQuery({
    variables: {args},
    pollInterval: REPO_POLL_INTERVAL_MS,
  });

  return {
    repo: data?.repo,
    error,
    loading,
  };
};

export default useRepoWithLinkedPipeline;
