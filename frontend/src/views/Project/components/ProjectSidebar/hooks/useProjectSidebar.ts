import {useCallback} from 'react';
import {useHistory, useRouteMatch} from 'react-router';

import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  projectRoute,
  lineageRoute,
} from '@dash-frontend/views/Project/utils/routes';

import {LINEAGE_PATH} from '../../../constants/projectPaths';

const useProjectSidebar = () => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const {sidebarSize, overlay} = useSidebarInfo();
  const {jobSets, loading: jobSetsLoading, error} = useJobSets({projectId});
  const lineageMatch = useRouteMatch({
    path: LINEAGE_PATH,
  });

  const handleClose = useCallback(() => {
    browserHistory.push(
      lineageMatch ? lineageRoute({projectId}) : projectRoute({projectId}),
    );
  }, [browserHistory, lineageMatch, projectId]);

  return {
    projectId,
    handleClose,
    sidebarSize,
    overlay,
    jobSets,
    jobSetsLoading,
    error,
  };
};

export default useProjectSidebar;
