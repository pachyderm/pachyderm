import useProject from './useProject';
import useUrlState from './useUrlState';

const useCurrentProject = () => {
  const {projectId} = useUrlState();
  const {project, loading} = useProject({id: projectId});

  return {
    currentProject: project,
    loading,
  };
};

export default useCurrentProject;
