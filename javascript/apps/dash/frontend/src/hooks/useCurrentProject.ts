import useProject from './useProject';
import useProjectParams from './useProjectParams';

const useCurrentProject = () => {
  const {projectId} = useProjectParams();
  const {project, loading} = useProject({id: projectId});

  return {
    currentProject: project,
    loading,
  };
};

export default useCurrentProject;
