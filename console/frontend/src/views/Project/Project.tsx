import React from 'react';
import Favicon from 'react-favicon';
import {Route, Redirect} from 'react-router';

import BrandedTitle from '@dash-frontend/components/BrandedTitle';
import {useEnterpriseActive} from '@dash-frontend/hooks/useEnterpriseActive';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {projectReposRoute} from '@dash-frontend/views/Project/utils/routes';
import {useNotificationBanner} from '@pachyderm/components';

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
  PROJECT_CONFIG_PATH,
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
  LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
  PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
  PROJECT_SIDENAV_PATHS,
} from './constants/projectPaths';
import styles from './Project.module.css';
import ProjectConfig from './ProjectConfig';

const Project: React.FC = () => {
  const {loading, enterpriseActive} = useEnterpriseActive();
  const {add} = useNotificationBanner();
  const {projectId} = useUrlState();

  return (
    <>
      <BrandedTitle title="Project" />
      {/* Safari will not progamatically update its icon. So we serve HPE
      favicon from index.html and set the pachyderm icon when we know we
      are not in enterprise. */}
      {!loading && !enterpriseActive && <Favicon url="/img/pachyderm.ico" />}
      <ProjectHeader />
      <div className={styles.view}>
        <Route path={PROJECT_CONFIG_PATH} exact>
          <ProjectConfig triggerNotification={add} />
        </Route>
        <Route path={PROJECT_SIDENAV_PATHS} exact>
          <ProjectSideNav />
        </Route>
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
            LINEAGE_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
            PROJECT_PIPELINE_LOGS_VIEWER_JOB_PATH_LATEST,
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
      </div>
    </>
  );
};

export default Project;
