import {useGetJobsQuery} from '@dash-frontend/generated/hooks';

export const useJobs = (projectId = '') => {
  const {data, error, loading} = useGetJobsQuery({
    variables: {args: {projectId}},
  });

  return {
    error,
    jobs: data?.jobs || [],
    loading,
  };
};
