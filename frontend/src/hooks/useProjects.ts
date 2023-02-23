import {PROJECTS_POLL_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import {useProjectsQuery} from '@dash-frontend/generated/hooks';

export const useProjects = () => {
  const {data, error, loading, refetch} = useProjectsQuery({
    pollInterval: PROJECTS_POLL_INTERVAL_MS,
  });

  return {
    error,
    projects: data?.projects || [],
    loading,
    refetch,
  };
};
