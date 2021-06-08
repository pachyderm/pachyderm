import {usePipelineJobsQuery} from '@dash-frontend/generated/hooks';
import {PipelineJobsQueryArgs} from '@graphqlTypes';

export const usePipelineJobs = (args: PipelineJobsQueryArgs) => {
  const {data, error, loading} = usePipelineJobsQuery({
    variables: {args},
    pollInterval: 5000,
  });

  return {
    error,
    pipelineJobs: data?.pipelineJobs || [],
    loading,
  };
};
