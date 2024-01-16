/* eslint-disable no-useless-escape */
/* eslint-disable prettier/prettier */
export const PROJECT_PATH = '/project/:projectId';
export const PROJECT_REPOS_PATH = '/project/:projectId/repos/:tabId?';
export const PROJECT_PIPELINES_PATH = '/project/:projectId/pipelines/:tabId?';
export const PROJECT_JOBS_PATH = '/project/:projectId/jobs/:tabId?';
export const PROJECT_RUNS_JOB_PATH = '/project/:projectId/jobs/subjobs?';
export const PROJECT_FILE_UPLOAD_PATH =
  '/project/:projectId/repos/:repoId/upload';
export const PROJECT_FILE_BROWSER_PATH =
  '/project/:projectId/repos/:repoId/commit/:commitId/:filePath?/';
export const PROJECT_FILE_BROWSER_PATH_LATEST =
  '/project/:projectId/repos/:repoId/latest';

export const PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST = `/project/:projectId/logs/pipelines/:pipelineId/logs`;
export const PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH = `/project/:projectId/pipelines/:pipelineId/jobs/:jobId/logs`;
export const PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH = `/project/:projectId/pipelines/:pipelineId/jobs/:jobId/logs/datum/:datumId?`;
export const PROJECT_JOB_LOGS_VIEWER_JOB_PATH = `/project/:projectId/jobs/:jobId/pipeline/:pipelineId/logs`;
export const PROJECT_JOB_LOGS_VIEWER_DATUM_PATH = `/project/:projectId/jobs/:jobId/pipeline/:pipelineId/logs/datum/:datumId?`;

export const LINEAGE_PATH = '/lineage/:projectId';

export const LINEAGE_REPO_PATH = '/lineage/:projectId/repos/:repoId';
export const LINEAGE_PIPELINE_PATH =
  '/lineage/:projectId/pipelines/:pipelineId/:tabId?';
export const LINEAGE_FILE_UPLOAD_PATH =
  '/lineage/:projectId/repos/:repoId/upload';

export const LINEAGE_FILE_BROWSER_PATH =
  '/lineage/:projectId/repos/:repoId/commit/:commitId/:filePath?/';
export const LINEAGE_FILE_BROWSER_PATH_LATEST =
  '/lineage/:projectId/repos/:repoId/latest';

export const LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST = `/lineage/:projectId/pipelines/:pipelineId/logs`;
export const LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH = `/lineage/:projectId/pipelines/:pipelineId/jobs/:jobId/logs`;
export const LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH = `/lineage/:projectId/pipelines/:pipelineId/jobs/:jobId/logs/datum/:datumId?`;
export const LINEAGE_JOB_LOGS_VIEWER_JOB_PATH = `/lineage/:projectId/jobs/:jobId/pipeline/:pipelineId/logs`;
export const LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH = `/lineage/:projectId/jobs/:jobId/pipeline/:pipelineId/logs/datum/:datumId?`;

export const CLUSTER_CONFIG = `/cluster/defaults`;
export const PROJECT_CONFIG_PATH = `/project/:projectId/defaults`;
export const CREATE_PIPELINE_PATH = `/lineage/:projectId/create/pipeline`;
export const UPDATE_PIPELINE_PATH = `/lineage/:projectId/update/pipeline/:pipelineId`;
export const DUPLICATE_PIPELINE_PATH = `/lineage/:projectId/duplicate/pipeline/:pipelineId`;

export const PROJECT_SIDENAV_PATHS = [
  LINEAGE_PATH,
  LINEAGE_REPO_PATH,
  LINEAGE_PIPELINE_PATH,
  LINEAGE_FILE_UPLOAD_PATH,
  LINEAGE_FILE_BROWSER_PATH,
  LINEAGE_FILE_BROWSER_PATH_LATEST,
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH,
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
  PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  PROJECT_JOB_LOGS_VIEWER_JOB_PATH,
  PROJECT_JOB_LOGS_VIEWER_DATUM_PATH,
  PROJECT_PATH,
  PROJECT_REPOS_PATH,
  PROJECT_PIPELINES_PATH,
  PROJECT_JOBS_PATH,
  PROJECT_RUNS_JOB_PATH,
  PROJECT_FILE_UPLOAD_PATH,
  PROJECT_FILE_BROWSER_PATH,
  PROJECT_FILE_BROWSER_PATH_LATEST,
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
  LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_JOB_LOGS_VIEWER_JOB_PATH,
  LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH,
];

export const PROJECT_PATHS = [
  PROJECT_CONFIG_PATH,
  CREATE_PIPELINE_PATH,
  UPDATE_PIPELINE_PATH,
  DUPLICATE_PIPELINE_PATH,
  ...PROJECT_SIDENAV_PATHS,
];
