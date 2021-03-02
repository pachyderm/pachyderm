import {useQuery} from '@apollo/client';

import {Project} from '@graphqlTypes';
import {GET_PROJECTS_QUERY} from 'queries/GetProjectsQuery';

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
