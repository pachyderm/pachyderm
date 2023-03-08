import {useMemo} from 'react';
import {useRouteMatch} from 'react-router';

import useLocalProjectSettings from '@dash-frontend/hooks/useLocalProjectSettings';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  LINEAGE_PIPELINE_PATH,
  LINEAGE_REPO_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';

export const DEFAULT_SIDEBAR_SIZE = 392;

const useSidebarInfo = () => {
  const match = useRouteMatch([LINEAGE_REPO_PATH, LINEAGE_PIPELINE_PATH]);
  const {projectId} = useUrlState();
  const [sidebarWidthSetting, handleUpdateSidebarWidth] =
    useLocalProjectSettings({
      projectId,
      key: 'sidebar_width',
    });

  const sidebarSize = useMemo(() => {
    return Number(sidebarWidthSetting || DEFAULT_SIDEBAR_SIZE);
  }, [sidebarWidthSetting]);

  return {
    sidebarSize,
    isOpen: Boolean(match),
    handleUpdateSidebarWidth,
  };
};

export default useSidebarInfo;
