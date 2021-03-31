import {useMemo} from 'react';
import {useParams} from 'react-router';

import {ProjectRouteParams} from '@dash-frontend/lib/types';

import {useProjects} from './useProjects';

const useCurrentProject = () => {
  const params = useParams<ProjectRouteParams>();
  const {projects, loading} = useProjects();

  const currentProject = useMemo(() => {
    return projects.find((project) => project.id === params.projectId);
  }, [params.projectId, projects]);

  return {
    currentProject,
    loading,
  };
};

export default useCurrentProject;
