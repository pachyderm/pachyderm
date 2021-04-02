import React from 'react';
import {Route, Switch} from 'react-router';

import JobList from '@dash-frontend/components/JobList';
import Sidebar from '@dash-frontend/components/Sidebar';

import useProjectSidebar from './hooks/useProjectSidebar';

const ProjectSidebar = () => {
  const {path, projectId, handleClose} = useProjectSidebar();

  return (
    <Sidebar overlay onClose={handleClose}>
      <Switch>
        <Route path={path} exact>
          <JobList projectId={projectId} expandActions />
        </Route>
      </Switch>
    </Sidebar>
  );
};

export default ProjectSidebar;
