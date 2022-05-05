import {useMemo} from 'react';
import {useRouteMatch} from 'react-router';

import {SidebarSize} from '@dash-frontend/lib/types';
import {
  LINEAGE_JOB_PATH,
  LINEAGE_JOBS_PATH,
  LINEAGE_PIPELINE_PATH,
  LINEAGE_REPO_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';

const useSidebarInfo = () => {
  const match = useRouteMatch([
    LINEAGE_REPO_PATH,
    LINEAGE_PIPELINE_PATH,
    LINEAGE_JOB_PATH,
    LINEAGE_JOBS_PATH,
  ]);

  const sidebarSize = useMemo<SidebarSize>(() => {
    switch (match?.path) {
      case LINEAGE_JOB_PATH:
      case LINEAGE_JOBS_PATH:
        return 'lg';
      default:
        return 'md';
    }
  }, [match]);

  return {
    sidebarSize,
    isOpen: Boolean(match),
    overlay: match?.path === LINEAGE_JOBS_PATH && match.isExact,
  };
};

export default useSidebarInfo;
