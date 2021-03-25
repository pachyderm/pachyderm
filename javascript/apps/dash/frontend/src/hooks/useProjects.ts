import {useProjectsQuery} from '@dash-frontend/generated/hooks';

export const useProjects = () => {
  const {data, error, loading} = useProjectsQuery();

  return {
    error,
    projects: data?.projects || [],
    loading,
  };
};
