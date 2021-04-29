import {generatePath} from 'react-router';

import {
  JOBS_PATH,
  PIPELINE_PATH,
  PROJECT_PATH,
  REPO_PATH,
  DAG_PATH,
} from '../constants/projectPaths';

export const projectRoute = ({projectId}: {projectId: string}) =>
  generatePath(PROJECT_PATH, {projectId: encodeURIComponent(projectId)});

export const dagRoute = ({
  projectId,
  dagId,
}: {
  projectId: string;
  dagId: string;
}) =>
  generatePath(DAG_PATH, {
    projectId: encodeURIComponent(projectId),
    dagId: encodeURIComponent(dagId),
  });

export const jobsRoute = ({projectId}: {projectId: string}) =>
  generatePath(JOBS_PATH, {
    projectId: encodeURIComponent(projectId),
  });

export const repoRoute = ({
  projectId,
  repoId,
  dagId,
  branchId,
}: {
  projectId: string;
  repoId: string;
  dagId: string;
  branchId: string;
}) =>
  generatePath(REPO_PATH, {
    projectId: encodeURIComponent(projectId),
    repoId: encodeURIComponent(repoId),
    dagId: encodeURIComponent(dagId),
    branchId: encodeURIComponent(branchId),
  });

export const pipelineRoute = ({
  projectId,
  pipelineId,
  tabId,
  dagId,
}: {
  projectId: string;
  pipelineId: string;
  tabId?: string;
  dagId: string;
}) =>
  generatePath(PIPELINE_PATH, {
    projectId: encodeURIComponent(projectId),
    dagId: encodeURIComponent(dagId),
    pipelineId: encodeURIComponent(pipelineId),
    tabId,
  });
