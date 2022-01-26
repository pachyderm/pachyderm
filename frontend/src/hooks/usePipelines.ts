import {PipelinesQueryArgs} from '@graphqlTypes';

import {usePipelinesQuery} from '@dash-frontend/generated/hooks';

const usePipelines = (args: PipelinesQueryArgs) => {
  const {data, error, loading} = usePipelinesQuery({
    variables: {args},
  });

  return {
    pipelines: data?.pipelines,
    error,
    loading,
  };
};

export default usePipelines;
