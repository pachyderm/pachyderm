import useRepoWithLinkedPipeline from './useRepoWithLinkedPipeline';
import useUrlState from './useUrlState';

const useCurrentRepoWithLinkedPipeline = () => {
  const {repoId, projectId} = useUrlState();
  const {repo, loading, error} = useRepoWithLinkedPipeline({
    id: repoId,
    projectId,
  });

  return {
    repo,
    error,
    loading,
  };
};

export default useCurrentRepoWithLinkedPipeline;
