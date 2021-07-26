import {useCallback} from 'react';
import {useHistory} from 'react-router';

import {useJobSets} from '@dash-frontend/hooks/useJobSets';
import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {projectRoute} from '@dash-frontend/views/Project/utils/routes';

const useProjectSidebar = () => {
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const {sidebarSize, overlay} = useSidebarInfo();
  const {jobSets, loading: jobSetsLoading} = useJobSets({projectId});

  const handleClose = useCallback(() => {
    browserHistory.push(projectRoute({projectId}));
  }, [browserHistory, projectId]);

  return {
    projectId,
    handleClose,
    sidebarSize,
    overlay,
    jobSets,
    jobSetsLoading,
  };
};

export default useProjectSidebar;
