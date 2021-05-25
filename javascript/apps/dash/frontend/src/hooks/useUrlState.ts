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
  const commitId = match?.params.commitId
    ? decodeURIComponent(match.params.commitId)
    : '';
  const filePath = match?.params.filePath
    ? decodeURIComponent(match.params.filePath)
    : '';

  return {branchId, projectId, repoId, pipelineId, commitId, filePath};
};

export default useUrlState;
