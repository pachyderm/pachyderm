import React from 'react';
import {Route, Redirect} from 'react-router';

import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {projectReposRoute} from '@dash-frontend/views/Project/utils/routes';

import {JobDatumViewer, PipelineDatumViewer} from '../DatumViewer';
import FileBrowser from '../FileBrowser';
import FileUpload from '../FileUpload';

import DAG from './components/DAG';
import JobSetList from './components/JobSetList';
import PipelineList from './components/PipelineList';
import ProjectHeader from './components/ProjectHeader';
import ProjectSideNav from './components/ProjectSideNav';
import RepoList from './components/RepoList';
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
  LINEAGE_FILE_BROWSER_PATH_LATEST,
  PROJECT_FILE_BROWSER_PATH_LATEST,
} from './constants/projectPaths';
import styles from './Project.module.css';

const Project: React.FC = () => {
  const {projectId} = useUrlState();

  return (
    <>
      <BrandedTitle title="Project" />
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
        <Route path={LINEAGE_PATH}>
          <DAG />
        </Route>
        <Route path={PROJECT_PIPELINES_PATH}>
          <div className={styles.listWrapper}>
            <PipelineList />
          </div>
        </Route>
        <Route path={PROJECT_REPOS_PATH}>
          <div className={styles.listWrapper}>
            <RepoList />
          </div>
        </Route>
        <Route path={PROJECT_JOBS_PATH}>
          <div className={styles.listWrapper}>
            <JobSetList />
          </div>
        </Route>
        <Route
          path={[
            LINEAGE_FILE_BROWSER_PATH,
            PROJECT_FILE_BROWSER_PATH,
            LINEAGE_FILE_BROWSER_PATH_LATEST,
            PROJECT_FILE_BROWSER_PATH_LATEST,
          ]}
        >
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
