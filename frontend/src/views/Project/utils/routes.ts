import {ExtractRouteParams, generatePath, matchPath} from 'react-router';

import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {generatePathWithSearch} from '@pachyderm/components';

import {
  LINEAGE_PATH,
  PROJECT_JOBS_PATH,
  PROJECT_RUNS_JOB_PATH,
  PROJECT_PATH,
  PROJECT_REPOS_PATH,
  PROJECT_PIPELINES_PATH,
  LINEAGE_FILE_BROWSER_PATH,
  PROJECT_FILE_BROWSER_PATH,
  PROJECT_FILE_PREVIEW_PATH,
  LINEAGE_REPO_PATH,
  LINEAGE_PIPELINE_PATH,
  PROJECT_FILE_UPLOAD_PATH,
  LINEAGE_FILE_UPLOAD_PATH,
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH,
  PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH,
  PROJECT_JOB_LOGS_VIEWER_JOB_PATH,
  LINEAGE_JOB_LOGS_VIEWER_JOB_PATH,
  PROJECT_JOB_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH,
  PROJECT_FILE_BROWSER_PATH_LATEST,
  LINEAGE_FILE_BROWSER_PATH_LATEST,
  LINEAGE_FILE_PREVIEW_PATH,
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
  CLUSTER_CONFIG,
  PROJECT_CONFIG_PATH,
  CREATE_PIPELINE_PATH,
  UPDATE_PIPELINE_PATH,
  DUPLICATE_PIPELINE_PATH,
} from '../constants/projectPaths';

const generateRouteFn = <S extends string>(path: S) => {
  return (params?: ExtractRouteParams<S>, withSearch = true) => {
    return withSearch
      ? generatePathWithSearch(path, params)
      : generatePath(path, params);
  };
};

const generateLineageOrProjectRouteFn = <S extends string>(
  projectPath: S,
  lineagePath: S,
) => {
  return (params?: ExtractRouteParams<S>, withSearch = true) => {
    if (matchPath(window.location.pathname, LINEAGE_PATH)) {
      return withSearch
        ? generatePathWithSearch(lineagePath, params)
        : generatePath(lineagePath, params);
    } else {
      return withSearch
        ? generatePathWithSearch(projectPath, params)
        : generatePath(projectPath, params);
    }
  };
};

export const clusterConfigRoute = CLUSTER_CONFIG;
export const projectConfigRoute = generateRouteFn(PROJECT_CONFIG_PATH);

export const lineageRoute = generateRouteFn(LINEAGE_PATH);
export const projectRoute = generateRouteFn(PROJECT_PATH);
export const projectJobsRoute = generateRouteFn(PROJECT_JOBS_PATH);
export const projectReposRoute = generateRouteFn(PROJECT_REPOS_PATH);
export const projectPipelinesRoute = generateRouteFn(PROJECT_PIPELINES_PATH);

export const jobRoute = generateRouteFn(PROJECT_RUNS_JOB_PATH);

export const jobsRoute = generateRouteFn(PROJECT_JOBS_PATH);
export const repoRoute = generateRouteFn(LINEAGE_REPO_PATH);
export const pipelineRoute = generateRouteFn(LINEAGE_PIPELINE_PATH);

export const createPipelineRoute = generateRouteFn(CREATE_PIPELINE_PATH);
export const updatePipelineRoute = generateRouteFn(UPDATE_PIPELINE_PATH);
export const duplicatePipelineRoute = generateRouteFn(DUPLICATE_PIPELINE_PATH);

export const fileBrowserRoute = generateLineageOrProjectRouteFn(
  PROJECT_FILE_BROWSER_PATH,
  LINEAGE_FILE_BROWSER_PATH,
);

export const fileBrowserLatestRoute = generateLineageOrProjectRouteFn(
  PROJECT_FILE_BROWSER_PATH_LATEST,
  LINEAGE_FILE_BROWSER_PATH_LATEST,
);

export const filePreviewRoute = generateLineageOrProjectRouteFn(
  PROJECT_FILE_PREVIEW_PATH,
  LINEAGE_FILE_PREVIEW_PATH,
);

const generateLineageOrProjectLogsRouteFn = <S extends string>(
  projectPipelinePath: S,
  projectJobPath: S,
  lineagePipelinePath: S,
  lineageJobPath: S,
) => {
  return (params?: ExtractRouteParams<S>, withSearch = true) => {
    const projectPath = matchPath(
      window.location.pathname,
      PROJECT_PIPELINES_PATH,
    )
      ? projectPipelinePath
      : projectJobPath;

    const lineagePath = matchPath(
      window.location.pathname,
      LINEAGE_PIPELINE_PATH,
    )
      ? lineagePipelinePath
      : lineageJobPath;

    if (matchPath(window.location.pathname, LINEAGE_PATH)) {
      return withSearch
        ? generatePathWithSearch(lineagePath, params)
        : encodeURI(generatePath(lineagePath, params));
    } else {
      return withSearch
        ? generatePathWithSearch(projectPath, params)
        : encodeURI(generatePath(projectPath, params));
    }
  };
};

export const logsViewerJobRoute = generateLineageOrProjectLogsRouteFn(
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH,
  PROJECT_JOB_LOGS_VIEWER_JOB_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH,
  LINEAGE_JOB_LOGS_VIEWER_JOB_PATH,
);
export const logsViewerDatumRoute = generateLineageOrProjectLogsRouteFn(
  PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  PROJECT_JOB_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH,
);

export const logsViewerLatestRoute = generateLineageOrProjectRouteFn(
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
);

export const fileUploadRoute = generateLineageOrProjectRouteFn(
  PROJECT_FILE_UPLOAD_PATH,
  LINEAGE_FILE_UPLOAD_PATH,
);

export const useSelectedRunRoute = ({
  projectId,
  jobId,
  pipelineId,
}: {
  projectId: string;
  jobId?: string;
  pipelineId?: string;
}) => {
  const {getUpdatedSearchParams} = useUrlQueryState();
  const newSearchParams = getUpdatedSearchParams({
    selectedJobs: jobId ? [jobId] : [],
    pipelineStep: pipelineId ? [pipelineId] : [],
  });

  return `${jobRoute({projectId}, false)}${newSearchParams}`;
};
