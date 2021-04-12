import {useRouteMatch} from 'react-router';

import {ProjectRouteParams} from '@dash-frontend/lib/types';
import {PROJECT_PATHS} from '@dash-frontend/views/Project/constants/projectPaths';

const useUrlState = () => {
  const match = useRouteMatch<ProjectRouteParams>({
    path: PROJECT_PATHS,
    exact: true,
  });

  const projectId =
    match && match.params.projectId
      ? decodeURIComponent(match.params.projectId)
      : '';
  const repoId =
    match && match.params.repoId ? decodeURIComponent(match.params.repoId) : '';
  const pipelineId =
    match && match.params.pipelineId
      ? decodeURIComponent(match.params.pipelineId)
      : '';

  return {projectId, repoId, pipelineId};
};

export default useUrlState;
