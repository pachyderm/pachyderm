import {useRouteMatch} from 'react-router';

import {ProjectRouteParams} from '@dash-frontend/lib/types';
import {PROJECT_PATHS} from '@dash-frontend/views/Project/constants/projectPaths';

const useUrlState = () => {
  const match = useRouteMatch<ProjectRouteParams>({
    path: PROJECT_PATHS,
    exact: true,
  });
  const projectId = match?.params.projectId
    ? decodeURIComponent(match.params.projectId)
    : '';
  const repoId = match?.params.repoId
    ? decodeURIComponent(match.params.repoId)
    : '';
  const pipelineId = match?.params.pipelineId
    ? decodeURIComponent(match.params.pipelineId)
    : '';
  const branchId = match?.params.branchId
    ? decodeURIComponent(match.params.branchId)
    : '';
  const dagId = match?.params.dagId
    ? decodeURIComponent(match.params.dagId)
    : '';

  return {branchId, dagId, projectId, repoId, pipelineId};
};

export default useUrlState;
