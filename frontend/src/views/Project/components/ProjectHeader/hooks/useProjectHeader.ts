import useCurrentProject from '@dash-frontend/hooks/useCurrentProject';

const useProjectHeader = () => {
  const {currentProject, loading} = useCurrentProject();

  return {
    projectName: currentProject?.name || '',
    loading,
  };
};

export default useProjectHeader;
