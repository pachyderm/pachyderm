import {useQuery} from '@apollo/client';

import {GET_PROJECTS_QUERY} from '@dash-frontend/queries/GetProjectsQuery';
import {Project} from '@graphqlTypes';

type ProjectsQueryResponse = {
  projects: Project[];
};

export const useProjects = () => {
  const {data, error, loading} = useQuery<ProjectsQueryResponse>(
    GET_PROJECTS_QUERY,
  );

  return {
    error,
    projects: data?.projects || [],
    loading,
  };
};
