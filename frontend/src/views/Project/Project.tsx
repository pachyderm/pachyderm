import {LoadingDots} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';
import {Helmet} from 'react-helmet';
import {Route} from 'react-router';

import View from '@dash-frontend/components/View';

import FileBrowser from '../FileBrowser';
import JobLogsViewer from '../LogsViewers/JobLogsViewer/JobLogsViewer';
import PipelineLogsViewer from '../LogsViewers/PipelineLogsViewer';

import DAGView from './components/DAGView';
import ProjectHeader from './components/ProjectHeader';
import ProjectSidebar from './components/ProjectSidebar';
import {NODE_HEIGHT, NODE_WIDTH} from './constants/nodeSizes';
import {
  FILE_BROWSER_PATH,
  LOGS_VIEWER_JOB_PATH,
  LOGS_VIEWER_PIPELINE_PATH,
} from './constants/projectPaths';
import {useProjectView} from './hooks/useProjectView';
import styles from './Project.module.css';
import ProjectTutorial from './tutorials/ProjectTutorial';

const Project: React.FC = () => {
  const {dags, loading, error, isSidebarOpen, sidebarSize} = useProjectView(
    NODE_WIDTH,
    NODE_HEIGHT,
  );

  return (
    <>
      <Helmet>
        <title>Project - Pachyderm Console</title>
      </Helmet>
      <ProjectHeader />
      <View className={styles.view}>
        {loading ? (
          <div
            className={classnames(styles.loadingContainer, {
              [styles.isSidebarOpen]: isSidebarOpen,
              [styles[sidebarSize]]: true,
            })}
          >
            <LoadingDots />
          </div>
        ) : (
          <DAGView dags={dags} loading={loading} error={error} />
        )}
        <ProjectSidebar />
        <Route path={FILE_BROWSER_PATH}>
          <FileBrowser />
        </Route>
        <Route path={LOGS_VIEWER_PIPELINE_PATH}>
          <PipelineLogsViewer />
        </Route>
        <Route path={LOGS_VIEWER_JOB_PATH}>
          <JobLogsViewer />
        </Route>
        <ProjectTutorial />
      </View>
    </>
  );
};

export default Project;
