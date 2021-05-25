import React from 'react';
import {Route, Switch} from 'react-router';

import JobList from '@dash-frontend/components/JobList';
import Sidebar from '@dash-frontend/components/Sidebar';

import {
  JOBS_PATH,
  PIPELINE_PATH,
  REPO_PATH,
} from '../../constants/projectPaths';

import PipelineDetails from './components/PipelineDetails';
import RepoDetails from './components/RepoDetails';
import useProjectSidebar from './hooks/useProjectSidebar';

const ProjectSidebar = () => {
  const {projectId, handleClose} = useProjectSidebar();

  return (
    <Route path={[JOBS_PATH, REPO_PATH, PIPELINE_PATH]}>
      <Sidebar
        overlay
        onClose={handleClose}
        data-testid={'ProjectSidebar__sidebar'}
      >
        <Switch>
          <Route path={JOBS_PATH} exact>
            <JobList projectId={projectId} expandActions showStatusFilter />
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
