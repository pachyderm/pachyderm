import {useCallback, useMemo} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {SidebarSize} from '@dash-frontend/lib/types';
import {
  JOBS_PATH,
  JOB_PATH,
  PIPELINE_PATH,
  REPO_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {projectRoute} from '@dash-frontend/views/Project/utils/routes';

const useProjectSidebar = () => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const match = useRouteMatch([JOBS_PATH, JOB_PATH, REPO_PATH, PIPELINE_PATH]);

  const sidebarSize = useMemo<SidebarSize>(() => {
    switch (match?.path) {
      case JOBS_PATH:
      case JOB_PATH:
        return 'lg';
      default:
        return 'md';
    }
  }, [match]);

  const handleClose = useCallback(() => {
    browserHistory.push(projectRoute({projectId}));
  }, [browserHistory, projectId]);

  return {
    projectId,
    handleClose,
    sidebarSize,
  };
};

export default useProjectSidebar;
