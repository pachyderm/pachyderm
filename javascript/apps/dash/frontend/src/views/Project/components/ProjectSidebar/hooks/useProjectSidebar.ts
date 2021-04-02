import {useCallback} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import useProjectParams from '@dash-frontend/hooks/useProjectParams';

const useProjectSidebar = () => {
  const {projectId} = useProjectParams();
  const browserHistory = useHistory();
  const {path} = useRouteMatch();

  const handleClose = useCallback(() => {
    browserHistory.push(`/project/${projectId}`);
  }, [browserHistory, projectId]);

  return {
    projectId,
    path,
    handleClose,
  };
};

export default useProjectSidebar;
