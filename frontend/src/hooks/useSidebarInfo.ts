import {useMemo} from 'react';
import {useRouteMatch} from 'react-router';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  LINEAGE_PIPELINE_PATH,
  LINEAGE_REPO_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';

export const DEFAULT_SIDEBAR_SIZE = 392;

const useSidebarInfo = () => {
  const match = useRouteMatch([LINEAGE_REPO_PATH, LINEAGE_PIPELINE_PATH]);
  const {projectId} = useUrlState();
  const {viewState} = useUrlQueryState();
  const [sidebarWidthSetting] = useLocalProjectSettings({
    projectId,
    key: 'sidebar_width',
  });

  const sidebarSize = useMemo(() => {
    return Number(
      viewState.sidebarWidth || sidebarWidthSetting || DEFAULT_SIDEBAR_SIZE,
    );
  }, [sidebarWidthSetting, viewState.sidebarWidth]);

  return {
    sidebarSize,
    isOpen: Boolean(match),
  };
};

export default useSidebarInfo;
