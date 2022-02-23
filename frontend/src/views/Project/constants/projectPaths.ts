/* eslint-disable no-useless-escape */
/* eslint-disable prettier/prettier */
export const PROJECT_PATH = '/project/:projectId';
export const PROJECT_REPOS_PATH = '/project/:projectId/repos';
export const PROJECT_PIPELINES_PATH = '/project/:projectId/pipelines';

export const PROJECT_REPO_PATH = '/project/:projectId/repos/:repoId/branch/:branchId';
export const PROJECT_PIPELINE_PATH = '/project/:projectId/pipelines/:pipelineId/:tabId?';
export const PROJECT_JOBS_PATH = '/project/:projectId/jobs';
export const PROJECT_JOB_PATH = '/project/:projectId/jobs/:jobId/:pipelineId?';
export const PROJECT_PIPELINE_JOB_PATH = '/project/:projectId/jobs/:jobId/:pipelineId';
export const PROJECT_FILE_UPLOAD_PATH = '/project/:projectId/repos/:repoId/upload';

export const LINEAGE_PATH = '/lineage/:projectId'

export const LINEAGE_REPO_PATH = '/lineage/:projectId/repos/:repoId/branch/:branchId';;
export const LINEAGE_PIPELINE_PATH = '/lineage/:projectId/pipelines/:pipelineId/:tabId?';
export const LINEAGE_JOBS_PATH = '/lineage/:projectId/jobs';
export const LINEAGE_JOB_PATH = '/lineage/:projectId/jobs/:jobId/:pipelineId?';
export const LINEAGE_PIPELINE_JOB_PATH = '/lineage/:projectId/jobs/:jobId/:pipelineId';
export const LINEAGE_FILE_UPLOAD_PATH = '/lineage/:projectId/repos/:repoId/upload';

export const FILE_BROWSER_PATH = '/project/:projectId/repos/:repoId/branch/:branchId/commit/:commitId/:filePath?';
export const FILE_BROWSER_DIR_PATH = `/project/:projectId/repos/:repoId/branch/:branchId/commit/:commitId/:filePath([^.]*)?`;
export const FILE_BROWSER_FILE_PATH = `/project/:projectId/repos/:repoId/branch/:branchId/commit/:commitId/:filePath(.+\.)`;

export const LOGS_VIEWER_PIPELINE_PATH = `/project/:projectId/pipelines/:pipelineId/logs`;
export const LOGS_VIEWER_JOB_PATH = `/project/:projectId/jobs/:jobId/:pipelineId/logs`;

export const TUTORIAL_PATH = [LINEAGE_PATH, FILE_BROWSER_PATH];

export const PROJECT_PATHS = [
  LINEAGE_PATH,
  LINEAGE_JOB_PATH,
  LINEAGE_JOBS_PATH,
  LINEAGE_PIPELINE_JOB_PATH,
  LINEAGE_REPO_PATH,
  LINEAGE_PIPELINE_PATH,
  LINEAGE_FILE_UPLOAD_PATH,
  PROJECT_PATH,
  PROJECT_REPOS_PATH,
  PROJECT_PIPELINES_PATH,
  PROJECT_JOBS_PATH,
  PROJECT_JOB_PATH,
  PROJECT_PIPELINE_JOB_PATH,
  PROJECT_REPO_PATH,
  PROJECT_PIPELINE_PATH,
  PROJECT_FILE_UPLOAD_PATH,
  FILE_BROWSER_PATH,
  FILE_BROWSER_DIR_PATH,
  FILE_BROWSER_FILE_PATH,
  LOGS_VIEWER_PIPELINE_PATH,
  LOGS_VIEWER_JOB_PATH,
];
