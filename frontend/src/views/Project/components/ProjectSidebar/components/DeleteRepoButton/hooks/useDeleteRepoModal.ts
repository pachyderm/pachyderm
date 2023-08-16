import {useHistory, useRouteMatch} from 'react-router';

import {useDeleteRepo} from '@dash-frontend/hooks/useDeleteRepo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {PROJECT_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {
  lineageRoute,
  projectReposRoute,
} from '@dash-frontend/views/Project/utils/routes';

const useDeleteRepoButton = () => {
  const {projectId, repoId} = useUrlState();
  const browserHistory = useHistory();
  const projectMatch = !!useRouteMatch(PROJECT_PATH);

  const onCompleted = () => {
    if (projectMatch) {
      browserHistory.push(projectReposRoute({projectId}));
    } else {
      browserHistory.push(lineageRoute({projectId}));
    }
  };

  const {deleteRepo, loading: updating, error} = useDeleteRepo(onCompleted);

  const onDelete = async () => {
    if (repoId) {
      try {
        await deleteRepo({
          variables: {
            args: {
              repo: {
                name: repoId,
              },
              projectId,
            },
          },
        });
      } catch (e) {
        return;
      }
    }
  };

  return {
    onDelete,
    updating,
    error,
  };
};

export default useDeleteRepoButton;
