/* eslint-disable no-useless-escape */
export const PROJECT_PATH = '/project/:projectId';
export const JOBS_PATH = '/project/:projectId/jobs';
export const JOB_PATH = '/project/:projectId/jobs/:jobId/:pipelineId?';
export const PIPELINE_JOB_PATH = '/project/:projectId/jobs/:jobId/:pipelineId';
export const REPO_PATH = '/project/:projectId/repo/:repoId/branch/:branchId';
export const PIPELINE_PATH = '/project/:projectId/pipeline/:pipelineId/:tabId?';
export const FILE_BROWSER_PATH =
  '/project/:projectId/repo/:repoId/branch/:branchId/commit/:commitId/:filePath?';
export const FILE_BROWSER_DIR_PATH = `/project/:projectId/repo/:repoId/branch/:branchId/commit/:commitId/:filePath([^.]*)?`;
export const FILE_BROWSER_FILE_PATH = `/project/:projectId/repo/:repoId/branch/:branchId/commit/:commitId/:filePath(.+\.)`;

export const LOGS_VIEWER_PIPELINE_PATH = `/project/:projectId/pipeline/:pipelineId/logs`;
export const LOGS_VIEWER_JOB_PATH = `/project/:projectId/jobs/:jobId/:pipelineId/logs`;

export const PROJECT_PATHS = [
  PROJECT_PATH,
  JOBS_PATH,
  JOB_PATH,
  PIPELINE_JOB_PATH,
  REPO_PATH,
  PIPELINE_PATH,
  FILE_BROWSER_PATH,
  FILE_BROWSER_DIR_PATH,
  FILE_BROWSER_FILE_PATH,
  LOGS_VIEWER_PIPELINE_PATH,
  LOGS_VIEWER_JOB_PATH,
];
