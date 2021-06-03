import React from 'react';
import {Route, Switch} from 'react-router';

import JobList from '@dash-frontend/components/JobList';
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

const emptyStateTitle = "Let's Start :)";
const emptyJobListMessage =
  'Create your first job! If there are any pipeline errors, fix those before you create a job.';

const ProjectSidebar = () => {
  const {projectId, handleClose, sidebarSize} = useProjectSidebar();

  return (
    <Route path={[JOBS_PATH, JOB_PATH, REPO_PATH, PIPELINE_PATH]}>
      <Sidebar
        overlay
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
              emptyStateTitle={emptyStateTitle}
              emptyStateMessage={emptyJobListMessage}
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

export default ProjectSidebar;
