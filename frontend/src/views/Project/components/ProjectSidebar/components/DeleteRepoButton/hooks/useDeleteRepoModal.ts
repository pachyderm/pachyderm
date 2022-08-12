import {useDeleteRepo} from '@dash-frontend/hooks/useDeleteRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const useDeleteRepoButton = () => {
  const {projectId, repoId} = useUrlState();
  const {deleteRepo, loading: updating, error} = useDeleteRepo(repoId);

  const onDelete = () => {
    if (repoId) {
      deleteRepo({
        variables: {
          args: {
            repo: {
              name: repoId,
            },
            projectId,
          },
        },
      });
    }
  };

  return {
    onDelete,
    updating,
    error,
  };
};

export default useDeleteRepoButton;
