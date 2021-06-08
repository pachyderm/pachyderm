import {useJobsQuery} from '@dash-frontend/generated/hooks';
import {JobsQueryArgs} from '@graphqlTypes';

export const useJobs = (args: JobsQueryArgs) => {
  const {data, error, loading} = useJobsQuery({
    variables: {args},
    pollInterval: 5000,
  });

  return {
    error,
    jobs: data?.jobs || [],
    loading,
  };
};
