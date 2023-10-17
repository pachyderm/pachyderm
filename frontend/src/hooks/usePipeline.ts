import {QueryFunctionOptions} from '@apollo/client';
import {PipelineQueryArgs} from '@graphqlTypes';

import {usePipelineQuery} from '@dash-frontend/generated/hooks';

const usePipeline = (args: PipelineQueryArgs, opts?: QueryFunctionOptions) => {
  const {data, error, loading} = usePipelineQuery({
    variables: {args},
    fetchPolicy: opts?.fetchPolicy,
    skip: opts?.skip,
  });

  return {
    pipeline: data?.pipeline,
    error,
    loading,
  };
};

export default usePipeline;
