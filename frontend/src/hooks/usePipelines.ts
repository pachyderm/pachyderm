import {PipelinesQueryArgs} from '@graphqlTypes';

import {PIPELINES_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {usePipelinesQuery} from '@dash-frontend/generated/hooks';

const usePipelines = (args: PipelinesQueryArgs) => {
  const {data, error, loading} = usePipelinesQuery({
    variables: {args},
    pollInterval: PIPELINES_POLL_INTERVAL_MS,
  });

  return {
    pipelines: data?.pipelines,
    error,
    loading,
  };
};

export default usePipelines;
