import useCurrentProject from '@dash-frontend/hooks/useCurrentProject';

const useProjectHeader = () => {
  const {currentProject, loading, error} = useCurrentProject();

  return {
    projectName: currentProject?.name || '',
    loading,
    error,
  };
};

export default useProjectHeader;
