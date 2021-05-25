import {generatePath} from 'react-router';

import {ProjectRouteParams} from '@dash-frontend/lib/types';

import {
  JOBS_PATH,
  PIPELINE_PATH,
  PROJECT_PATH,
  REPO_PATH,
  FILE_BROWSER_PATH,
} from '../constants/projectPaths';

const generatePathWithSearch = (
  pathTemplate: string,
  params: ProjectRouteParams,
) => {
  const path = generatePath(pathTemplate, {...params});
  const urlParams = new URLSearchParams(window.location.search);
  return path + '?' + urlParams.toString();
};

export const projectRoute = ({
  projectId,
  withSearch = true,
}: {
  projectId: string;
  withSearch?: boolean;
}) => {
  return withSearch
    ? generatePathWithSearch(PROJECT_PATH, {
        projectId: encodeURIComponent(projectId),
      })
    : generatePath(PROJECT_PATH, {
        projectId: encodeURIComponent(projectId),
      });
};

export const jobsRoute = ({projectId}: {projectId: string}) =>
  generatePathWithSearch(JOBS_PATH, {
    projectId: encodeURIComponent(projectId),
  });

export const repoRoute = ({
  projectId,
  repoId,
  branchId,
}: {
  projectId: string;
  repoId: string;
  branchId: string;
}) =>
  generatePathWithSearch(REPO_PATH, {
    projectId: encodeURIComponent(projectId),
    repoId: encodeURIComponent(repoId),
    branchId: encodeURIComponent(branchId),
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
  generatePathWithSearch(PIPELINE_PATH, {
    projectId: encodeURIComponent(projectId),
    pipelineId: encodeURIComponent(pipelineId),
    tabId,
  });

export const fileBrowserRoute = ({
  projectId,
  repoId,
  commitId,
  filePath,
  branchId,
}: {
  projectId: string;
  repoId: string;
  commitId: string;
  filePath?: string;
  branchId: string;
}) =>
  generatePathWithSearch(FILE_BROWSER_PATH, {
    projectId: encodeURIComponent(projectId),
    repoId: encodeURIComponent(repoId),
    commitId: encodeURIComponent(commitId),
    branchId: encodeURIComponent(branchId),
    filePath: filePath ? encodeURIComponent(filePath) : undefined,
  });
