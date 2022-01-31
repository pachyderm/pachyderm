import {useMemo} from 'react';
import {useRouteMatch} from 'react-router';

import {SidebarSize} from '@dash-frontend/lib/types';
import {
  PROJECT_JOBS_PATH,
  PROJECT_JOB_PATH,
  PROJECT_PIPELINE_PATH,
  PROJECT_REPO_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';

const useSidebarInfo = () => {
  const match = useRouteMatch([
    PROJECT_JOBS_PATH,
    PROJECT_JOB_PATH,
    PROJECT_REPO_PATH,
    PROJECT_PIPELINE_PATH,
  ]);

  const sidebarSize = useMemo<SidebarSize>(() => {
    switch (match?.path) {
      case PROJECT_JOBS_PATH:
      case PROJECT_JOB_PATH:
        return 'lg';
      default:
        return 'md';
    }
  }, [match]);

  return {
    sidebarSize,
    isOpen: Boolean(match),
    overlay: match?.path === PROJECT_JOBS_PATH && match.isExact,
  };
};

export default useSidebarInfo;
