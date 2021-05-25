export const PROJECT_PATH = '/project/:projectId';
export const JOBS_PATH = '/project/:projectId/jobs';
export const REPO_PATH = '/project/:projectId/repo/:repoId/branch/:branchId';
export const PIPELINE_PATH = '/project/:projectId/pipeline/:pipelineId/:tabId?';
export const FILE_BROWSER_PATH =
  '/project/:projectId/repo/:repoId/branch/:branchId/commit/:commitId/:filePath?';

export const PROJECT_PATHS = [
  PROJECT_PATH,
  JOBS_PATH,
  REPO_PATH,
  PIPELINE_PATH,
  FILE_BROWSER_PATH,
];
