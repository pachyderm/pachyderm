import {useGetJobsQuery} from '@dash-frontend/generated/hooks';
import {GetJobsQueryVariables} from '@graphqlTypes';

export const useJobs = (args: GetJobsQueryVariables['args']) => {
  const {data, error, loading} = useGetJobsQuery({
    variables: {args},
  });

  return {
    error,
    jobs: data?.jobs || [],
    loading,
  };
};
