export const PROJECT_PATH = '/project/:projectId';
export const DAG_PATH = '/project/:projectId/dag/:dagId/';
export const JOBS_PATH = '/project/:projectId/jobs';
export const REPO_PATH =
  '/project/:projectId/dag/:dagId/repo/:repoId/branch/:branchId';
export const PIPELINE_PATH =
  '/project/:projectId/dag/:dagId/pipeline/:pipelineId/:tabId?';

export const PROJECT_PATHS = [
  PROJECT_PATH,
  JOBS_PATH,
  REPO_PATH,
  PIPELINE_PATH,
  DAG_PATH,
];
