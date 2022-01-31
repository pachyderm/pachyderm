import {useRouteMatch} from 'react-router';

import {LINEAGE_JOB_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

const useIsViewingJob = () => {
  const isViewingJob = useRouteMatch(LINEAGE_JOB_PATH);

  return Boolean(isViewingJob);
};

export default useIsViewingJob;
