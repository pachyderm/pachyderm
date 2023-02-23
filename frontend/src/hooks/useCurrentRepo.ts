import useRepo from './useRepo';
import useUrlState from './useUrlState';

const useCurrentRepo = () => {
  const {repoId, projectId} = useUrlState();
  const {repo, loading, error} = useRepo({
    id: repoId,
    projectId,
  });

  return {
    repo,
    error,
    loading,
  };
};

export default useCurrentRepo;
