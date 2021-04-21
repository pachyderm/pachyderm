import useRepo from './useRepo';
import useUrlState from './useUrlState';

const useCurrentRepo = () => {
  const {repoId, projectId} = useUrlState();
  const {repo, loading} = useRepo({
    id: repoId,
    projectId,
  });

  return {
    repo,
    loading,
  };
};

export default useCurrentRepo;
