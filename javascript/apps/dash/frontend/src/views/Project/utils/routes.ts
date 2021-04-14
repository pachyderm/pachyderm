import {generatePath} from 'react-router';

import {
  JOBS_PATH,
  PIPELINE_PATH,
  PROJECT_PATH,
  REPO_PATH,
} from '../constants/projectPaths';

export const projectRoute = ({projectId}: {projectId: string}) =>
  generatePath(PROJECT_PATH, {projectId: encodeURIComponent(projectId)});
export const jobsRoute = ({projectId}: {projectId: string}) =>
  generatePath(JOBS_PATH, {projectId: encodeURIComponent(projectId)});
export const repoRoute = ({
  projectId,
  repoId,
}: {
  projectId: string;
  repoId: string;
}) =>
  generatePath(REPO_PATH, {
    projectId: encodeURIComponent(projectId),
    repoId: encodeURIComponent(repoId),
  });
export const pipelineRoute = ({
  projectId,
  pipelineId,
  tabId,
}: {
  projectId: string;
  pipelineId: string;
  tabId?: string;
}) =>
  generatePath(PIPELINE_PATH, {
    projectId: encodeURIComponent(projectId),
    pipelineId: encodeURIComponent(pipelineId),
    tabId,
  });
