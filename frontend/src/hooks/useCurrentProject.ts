import {useProject} from './useProject';
import useUrlState from './useUrlState';

const useCurrentProject = () => {
  const {projectId} = useUrlState();
  const {project, loading, error} = useProject(projectId);

  return {
    projectId,
    currentProject: project,
    loading,
    error,
  };
};

export default useCurrentProject;
