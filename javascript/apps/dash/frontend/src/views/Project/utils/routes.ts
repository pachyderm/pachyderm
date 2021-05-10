import {generatePath} from 'react-router';

import {ProjectRouteParams} from '@dash-frontend/lib/types';

import {
  JOBS_PATH,
  PIPELINE_PATH,
  PROJECT_PATH,
  REPO_PATH,
  DAG_PATH,
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

export const dagRoute = ({
  projectId,
  dagId,
}: {
  projectId: string;
  dagId: string;
}) =>
  generatePathWithSearch(DAG_PATH, {
    projectId: encodeURIComponent(projectId),
    dagId: encodeURIComponent(dagId),
  });

export const jobsRoute = ({projectId}: {projectId: string}) =>
  generatePathWithSearch(JOBS_PATH, {
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
  generatePathWithSearch(REPO_PATH, {
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
  generatePathWithSearch(PIPELINE_PATH, {
    projectId: encodeURIComponent(projectId),
    dagId: encodeURIComponent(dagId),
    pipelineId: encodeURIComponent(pipelineId),
    tabId,
  });
