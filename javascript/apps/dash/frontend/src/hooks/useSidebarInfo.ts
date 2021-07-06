import {useMemo} from 'react';
import {useRouteMatch} from 'react-router';

import {SidebarSize} from '@dash-frontend/lib/types';
import {
  JOBS_PATH,
  JOB_PATH,
  PIPELINE_PATH,
  REPO_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';

const useSidebarInfo = () => {
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

  return {
    sidebarSize,
    isOpen: Boolean(match),
    overlay: match?.path === JOBS_PATH && match.isExact,
  };
};

export default useSidebarInfo;
