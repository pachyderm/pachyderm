import React from 'react';
import {Helmet} from 'react-helmet';
import {Route} from 'react-router';

import FileBrowser from '../FileBrowser';
import JobLogsViewer from '../LogsViewers/JobLogsViewer/JobLogsViewer';
import PipelineLogsViewer from '../LogsViewers/PipelineLogsViewer';

import ProjectDetails from './components/ProjectDetails';
import ProjectHeader from './components/ProjectHeader';
import ProjectJobList from './components/ProjectJobList';
import ProjectSidebar from './components/ProjectSidebar';
import ProjectSideNav from './components/ProjectSideNav';
import {
  PROJECT_JOBS_PATH,
  PROJECT_REPOS_PATH,
  PROJECT_PIPELINES_PATH,
  FILE_BROWSER_PATH,
  LOGS_VIEWER_JOB_PATH,
  LOGS_VIEWER_PIPELINE_PATH,
  LINEAGE_PATH,
} from './constants/projectPaths';
import styles from './Project.module.css';

const Project: React.FC = () => {
  return (
    <>
      <Helmet>
        <title>Project - Pachyderm Console</title>
      </Helmet>
      <ProjectHeader />
      <div className={styles.view}>
        <ProjectSideNav />
        <Route
          path={[PROJECT_REPOS_PATH, PROJECT_PIPELINES_PATH, LINEAGE_PATH]}
        >
          <ProjectDetails />
        </Route>
        <Route path={PROJECT_JOBS_PATH}>
          <ProjectJobList />
        </Route>
        <Route
          path={[PROJECT_REPOS_PATH, PROJECT_PIPELINES_PATH, PROJECT_JOBS_PATH]}
        >
          <ProjectSidebar resizable={false} />
        </Route>
        <Route path={FILE_BROWSER_PATH}>
          <FileBrowser />
        </Route>
        <Route path={LOGS_VIEWER_PIPELINE_PATH}>
          <PipelineLogsViewer />
        </Route>
        <Route path={LOGS_VIEWER_JOB_PATH}>
          <JobLogsViewer />
        </Route>
      </div>
    </>
  );
};

export default Project;
