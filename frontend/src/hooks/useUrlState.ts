import {match, useRouteMatch} from 'react-router';

import {ProjectRouteParams} from '@dash-frontend/lib/types';
import {PROJECT_PATHS} from '@dash-frontend/views/Project/constants/projectPaths';

const getDecodedRouteParam = (
  match: match<ProjectRouteParams> | null,
  key: keyof ProjectRouteParams,
) => {
  const param = match?.params[key];
  return param ? decodeURIComponent(param) : '';
};

const useUrlState = () => {
  const match = useRouteMatch<ProjectRouteParams>({
    path: PROJECT_PATHS,
    exact: true,
  });
  const projectId = getDecodedRouteParam(match, 'projectId');
  const repoId = getDecodedRouteParam(match, 'repoId');
  const pipelineId = getDecodedRouteParam(match, 'pipelineId');
  const commitId = getDecodedRouteParam(match, 'commitId');
  const filePath = getDecodedRouteParam(match, 'filePath');
  const jobId = getDecodedRouteParam(match, 'jobId');
  const datumId = getDecodedRouteParam(match, 'datumId');

  return {
    projectId,
    repoId,
    pipelineId,
    commitId,
    filePath,
    jobId,
    datumId,
  };
};

export default useUrlState;
