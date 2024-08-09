import {useCallback} from 'react';
import {useHistory} from 'react-router';

import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {lineageRoute} from '@dash-frontend/views/Project/utils/routes';

const useProjectSidebar = () => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const {sidebarSize} = useSidebarInfo();

  const handleClose = useCallback(() => {
    browserHistory.push(lineageRoute({projectId}));
  }, [browserHistory, projectId]);

  return {
    handleClose,
    sidebarSize,
  };
};

export default useProjectSidebar;
