import {useGetJobsQuery} from '@dash-frontend/generated/hooks';

export const useJobs = (projectId = '', limit?: number) => {
  const {data, error, loading} = useGetJobsQuery({
    variables: {args: {projectId, limit}},
  });

  return {
    error,
    jobs: data?.jobs || [],
    loading,
  };
};
