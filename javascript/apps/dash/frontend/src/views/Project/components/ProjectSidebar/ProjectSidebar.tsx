import React, {memo} from 'react';
import {Route, Switch} from 'react-router';

import JobList from '@dash-frontend/components/JobList';
import {
  CREATE_FIRST_JOB_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/ListEmptyState/constants/ListEmptyStateConstants';
import Sidebar from '@dash-frontend/components/Sidebar';

import {
  JOBS_PATH,
  PIPELINE_PATH,
  REPO_PATH,
  JOB_PATH,
} from '../../constants/projectPaths';

import JobDetails from './components/JobDetails';
import PipelineDetails from './components/PipelineDetails';
import RepoDetails from './components/RepoDetails';
import useProjectSidebar from './hooks/useProjectSidebar';

const ProjectSidebar = () => {
  const {projectId, handleClose, sidebarSize, overlay} = useProjectSidebar();

  return (
    <Route path={[JOBS_PATH, JOB_PATH, REPO_PATH, PIPELINE_PATH]}>
      <Sidebar
        fixed
        overlay={overlay}
        size={sidebarSize}
        onClose={handleClose}
        data-testid={'ProjectSidebar__sidebar'}
      >
        <Switch>
          <Route path={JOBS_PATH} exact>
            <JobList
              projectId={projectId}
              expandActions
              showStatusFilter
              emptyStateTitle={LETS_START_TITLE}
              emptyStateMessage={CREATE_FIRST_JOB_MESSAGE}
            />
          </Route>
          <Route path={JOB_PATH}>
            <JobDetails />
          </Route>
          <Route path={REPO_PATH}>
            <RepoDetails />
          </Route>
          <Route path={PIPELINE_PATH} exact>
            <PipelineDetails />
          </Route>
        </Switch>
      </Sidebar>
    </Route>
  );
};

export default memo(ProjectSidebar);
