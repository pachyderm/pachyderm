import React from 'react';
import {Route, Switch} from 'react-router';

import Sidebar from '@dash-frontend/components/Sidebar';
import {InputOutputNodesMap} from '@dash-frontend/lib/types';

import {
  LINEAGE_PIPELINE_PATH,
  LINEAGE_REPO_PATH,
} from '../../constants/projectPaths';

import PipelineDetails from './components/PipelineDetails';
import RepoDetails from './components/RepoDetails';
import useProjectSidebar from './hooks/useProjectSidebar';
import styles from './ProjectSidebar.module.css';

type ProjectSidebarProps = {
  inputOutputNodesMap?: InputOutputNodesMap;
  resizable?: boolean;
};

const ProjectSidebar: React.FC<ProjectSidebarProps> = ({
  inputOutputNodesMap,
  resizable = true,
}) => {
  const {handleClose} = useProjectSidebar();

  const resizableProps = resizable
    ? {
        fixed: true,
        onClose: handleClose,
        resizable: true,
      }
    : {};

  return (
    <div className={styles.base}>
      <Route path={[LINEAGE_REPO_PATH, LINEAGE_PIPELINE_PATH]}>
        <Sidebar data-testid="ProjectSidebar__sidebar" {...resizableProps}>
          <Switch>
            <Route path={LINEAGE_REPO_PATH}>
              <RepoDetails />
            </Route>
            <Route path={LINEAGE_PIPELINE_PATH} exact>
              <PipelineDetails inputOutputNodesMap={inputOutputNodesMap} />
            </Route>
          </Switch>
        </Sidebar>
      </Route>
    </div>
  );
};

export default ProjectSidebar;
