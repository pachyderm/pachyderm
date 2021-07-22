import {useRouteMatch} from 'react-router';

import {JOB_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

const useIsViewingJob = () => {
  const isViewingJob = useRouteMatch(JOB_PATH);

  return Boolean(isViewingJob);
};

export default useIsViewingJob;
