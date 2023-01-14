import React from 'react';
import {Helmet} from 'react-helmet';
import {Route, Redirect} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {projectReposRoute} from '@dash-frontend/views/Project/utils/routes';

import {JobDatumViewer, PipelineDatumViewer} from '../DatumViewer';
import FileBrowser from '../FileBrowser';
import FileUpload from '../FileUpload';

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
  LINEAGE_PATH,
  PROJECT_FILE_UPLOAD_PATH,
  LINEAGE_FILE_UPLOAD_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH,
  PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH,
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH,
  LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH,
  LINEAGE_JOB_LOGS_VIEWER_JOB_PATH,
  PROJECT_JOB_LOGS_VIEWER_DATUM_PATH,
  PROJECT_JOB_LOGS_VIEWER_JOB_PATH,
} from './constants/projectPaths';
import {useProjectView} from './hooks/useProjectView';
import styles from './Project.module.css';

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
            PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH,
            PROJECT_PIPELINE_LOGS_VIEWER_DATUM_PATH,
            LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH,
            LINEAGE_PIPELINE_LOGS_VIEWER_DATUM_PATH,
          ]}
        >
          <PipelineDatumViewer />
        </Route>
        <Route
          path={[
            PROJECT_JOB_LOGS_VIEWER_JOB_PATH,
            PROJECT_JOB_LOGS_VIEWER_DATUM_PATH,
            LINEAGE_JOB_LOGS_VIEWER_JOB_PATH,
            LINEAGE_JOB_LOGS_VIEWER_DATUM_PATH,
          ]}
        >
          <JobDatumViewer />
        </Route>

        <Route path={[PROJECT_FILE_UPLOAD_PATH, LINEAGE_FILE_UPLOAD_PATH]}>
          <FileUpload />
        </Route>
        {/* Tutorial is temporarily disabled because of "Project" Console Support */}
      </div>
    </>
  );
};

export default Project;
