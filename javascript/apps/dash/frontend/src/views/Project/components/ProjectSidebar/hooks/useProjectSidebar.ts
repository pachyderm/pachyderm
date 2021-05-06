import {useCallback} from 'react';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  dagRoute,
  projectRoute,
} from '@dash-frontend/views/Project/utils/routes';

const useProjectSidebar = () => {
  const {projectId, dagId} = useUrlState();
  const browserHistory = useHistory();

  const handleClose = useCallback(() => {
    browserHistory.push(
      dagId ? dagRoute({projectId, dagId}) : projectRoute({projectId}),
    );
  }, [browserHistory, projectId, dagId]);

  return {
    projectId,
    handleClose,
  };
};

export default useProjectSidebar;
