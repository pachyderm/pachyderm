import {usePipelineQuery} from '@dash-frontend/generated/hooks';
import {PipelineQueryArgs} from '@graphqlTypes';

const usePipeline = (args: PipelineQueryArgs) => {
  const {data, error, loading} = usePipelineQuery({
    variables: {args},
  });

  return {
    pipeline: data?.pipeline,
    error,
    loading,
  };
};

export default usePipeline;
