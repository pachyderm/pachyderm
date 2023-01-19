import {ExtractRouteParams, generatePath, matchPath} from 'react-router';

import {generatePathWithSearch} from '@pachyderm/components';

import {
  LINEAGE_PATH,
  PROJECT_JOBS_PATH,
  PROJECT_PIPELINE_PATH,
  PROJECT_PIPELINE_JOB_PATH,
  PROJECT_PATH,
  PROJECT_REPOS_PATH,
  PROJECT_PIPELINES_PATH,
  PROJECT_REPO_PATH,
  LINEAGE_FILE_BROWSER_PATH,
  PROJECT_FILE_BROWSER_PATH,
  PROJECT_JOB_PATH,
  LINEAGE_JOBS_PATH,
  LINEAGE_JOB_PATH,
  LINEAGE_PIPELINE_JOB_PATH,
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

export const lineageRoute = generateRouteFn(LINEAGE_PATH);
export const projectRoute = generateRouteFn(PROJECT_PATH);
export const projectReposRoute = generateRouteFn(PROJECT_REPOS_PATH);
export const projectPipelinesRoute = generateRouteFn(PROJECT_PIPELINES_PATH);

export const jobRoute = generateLineageOrProjectRouteFn(
  PROJECT_JOB_PATH,
  LINEAGE_JOB_PATH,
);
export const pipelineJobRoute = generateLineageOrProjectRouteFn(
  PROJECT_PIPELINE_JOB_PATH,
  LINEAGE_PIPELINE_JOB_PATH,
);
export const jobsRoute = generateLineageOrProjectRouteFn(
  PROJECT_JOBS_PATH,
  LINEAGE_JOBS_PATH,
);
export const repoRoute = generateLineageOrProjectRouteFn(
  PROJECT_REPO_PATH,
  LINEAGE_REPO_PATH,
);
export const pipelineRoute = generateLineageOrProjectRouteFn(
  PROJECT_PIPELINE_PATH,
  LINEAGE_PIPELINE_PATH,
);

export const fileBrowserRoute = generateLineageOrProjectRouteFn(
  PROJECT_FILE_BROWSER_PATH,
  LINEAGE_FILE_BROWSER_PATH,
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
      PROJECT_PIPELINE_PATH,
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

export const fileUploadRoute = generateLineageOrProjectRouteFn(
  PROJECT_FILE_UPLOAD_PATH,
  LINEAGE_FILE_UPLOAD_PATH,
);
