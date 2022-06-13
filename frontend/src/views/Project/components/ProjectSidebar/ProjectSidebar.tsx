import classnames from 'classnames';
import React from 'react';
import {Route, Switch} from 'react-router';

import {
  CREATE_FIRST_JOB_MESSAGE,
  LETS_START_TITLE,
} from '@dash-frontend/components/EmptyState/constants/EmptyStateConstants';
import JobList from '@dash-frontend/components/JobList';
import Sidebar from '@dash-frontend/components/Sidebar';

import {
  PROJECT_JOBS_PATH,
  PROJECT_PIPELINE_PATH,
  PROJECT_REPO_PATH,
  PROJECT_REPOS_PATH,
  PROJECT_PIPELINES_PATH,
  PROJECT_JOB_PATH,
  LINEAGE_JOBS_PATH,
  LINEAGE_REPO_PATH,
  LINEAGE_PIPELINE_PATH,
  LINEAGE_JOB_PATH,
} from '../../constants/projectPaths';

import JobDetails from './components/JobDetails';
import PipelineDetails from './components/PipelineDetails';
import RepoDetails from './components/RepoDetails';
import useProjectSidebar from './hooks/useProjectSidebar';
import styles from './ProjectSidebar.module.css';

type ProjectSidebarProps = {
  dagLinks?: Record<string, string[]>;
  dagsLoading?: boolean;
  resizable?: boolean;
};

const ProjectSidebar: React.FC<ProjectSidebarProps> = ({
  dagLinks,
  dagsLoading,
  resizable = true,
}) => {
  const {
    projectId,
    handleClose,
    sidebarSize,
    overlay,
    jobSets,
    jobSetsLoading,
  } = useProjectSidebar();

  const resizableProps = resizable
    ? {
        fixed: true,
        onClose: handleClose,
        overlay,
        defaultSize: sidebarSize,
        resizable: true,
      }
    : {};

  return (
    <div className={classnames(styles.base, {[styles.fixed]: !resizable})}>
      <Route
        path={[
          PROJECT_JOBS_PATH,
          PROJECT_JOB_PATH,
          PROJECT_REPO_PATH,
          PROJECT_REPOS_PATH,
          PROJECT_PIPELINES_PATH,
          LINEAGE_REPO_PATH,
          LINEAGE_JOB_PATH,
          LINEAGE_PIPELINE_PATH,
          PROJECT_PIPELINE_PATH,
          LINEAGE_JOBS_PATH,
        ]}
      >
        <Sidebar data-testid="ProjectSidebar__sidebar" {...resizableProps}>
          <Switch>
            <Route path={[LINEAGE_JOBS_PATH]} exact>
              <JobList
                projectId={projectId}
                jobs={jobSets}
                loading={jobSetsLoading}
                expandActions
                showStatusFilter
                emptyStateTitle={LETS_START_TITLE}
                emptyStateMessage={CREATE_FIRST_JOB_MESSAGE}
              />
            </Route>
            <Route path={[PROJECT_JOB_PATH, LINEAGE_JOB_PATH]}>
              <JobDetails />
            </Route>
            <Route path={[PROJECT_REPO_PATH, LINEAGE_REPO_PATH]}>
              <RepoDetails dagsLoading={!!dagsLoading} dagLinks={dagLinks} />
            </Route>
            <Route path={[PROJECT_PIPELINE_PATH, LINEAGE_PIPELINE_PATH]} exact>
              <PipelineDetails
                dagsLoading={!!dagsLoading}
                dagLinks={dagLinks}
              />
            </Route>
          </Switch>
        </Sidebar>
      </Route>
    </div>
  );
};

export default ProjectSidebar;
