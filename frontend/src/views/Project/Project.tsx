import React from 'react';
import {Helmet} from 'react-helmet';
import {Route, Redirect} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {projectReposRoute} from '@dash-frontend/views/Project/utils/routes';

import FileBrowser from '../FileBrowser';
import FileUpload from '../FileUpload';
import JobLogsViewer from '../LogsViewers/JobLogsViewer/JobLogsViewer';
import PipelineLogsViewer from '../LogsViewers/PipelineLogsViewer';

import ProjectDetails from './components/ProjectDetails';
import ProjectHeader from './components/ProjectHeader';
import ProjectJobList from './components/ProjectJobList';
import ProjectSideNav from './components/ProjectSideNav';
import {
  PROJECT_PATH,
  PROJECT_JOBS_PATH,
  PROJECT_REPOS_PATH,
  PROJECT_PIPELINES_PATH,
  LINEAGE_FILE_BROWSER_PATH,
  PROJECT_FILE_BROWSER_PATH,
  LINEAGE_LOGS_VIEWER_JOB_PATH,
  LINEAGE_LOGS_VIEWER_PIPELINE_PATH,
  PROJECT_LOGS_VIEWER_JOB_PATH,
  PROJECT_LOGS_VIEWER_PIPELINE_PATH,
  LINEAGE_PATH,
  PROJECT_FILE_UPLOAD_PATH,
  LINEAGE_FILE_UPLOAD_PATH,
} from './constants/projectPaths';
import {useProjectView} from './hooks/useProjectView';
import styles from './Project.module.css';
import ProjectTutorial from './tutorials/ProjectTutorial';

const Project: React.FC = () => {
  const {projectId} = useUrlState();

  const projectProps = useProjectView();

  return (
    <>
      <Helmet>
        <title>Project - Pachyderm Console</title>
      </Helmet>
      <ProjectHeader />
      <div className={styles.view}>
        <ProjectSideNav />
        <Route path={PROJECT_PATH} exact>
          <Redirect
            to={projectReposRoute({
              projectId,
            })}
          />
        </Route>
        <Route
          path={[PROJECT_REPOS_PATH, PROJECT_PIPELINES_PATH, LINEAGE_PATH]}
        >
          <ProjectDetails {...projectProps} />
        </Route>
        <Route path={PROJECT_JOBS_PATH}>
          <ProjectJobList />
        </Route>
        <Route path={[LINEAGE_FILE_BROWSER_PATH, PROJECT_FILE_BROWSER_PATH]}>
          <FileBrowser />
        </Route>
        <Route
          path={[
            PROJECT_LOGS_VIEWER_PIPELINE_PATH,
            LINEAGE_LOGS_VIEWER_PIPELINE_PATH,
          ]}
        >
          <PipelineLogsViewer />
        </Route>
        <Route
          path={[PROJECT_LOGS_VIEWER_JOB_PATH, LINEAGE_LOGS_VIEWER_JOB_PATH]}
        >
          <JobLogsViewer />
        </Route>
        <Route path={[PROJECT_FILE_UPLOAD_PATH, LINEAGE_FILE_UPLOAD_PATH]}>
          <FileUpload />
        </Route>
        <ProjectTutorial />
      </div>
    </>
  );
};

export default Project;
