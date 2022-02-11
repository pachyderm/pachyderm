import {LoadingDots} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';
import {Route} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';

import DAGView from '../../components/DAGView';
import ListView from '../../components/ListView';
import ProjectSidebar from '../../components/ProjectSidebar';
import {NODE_HEIGHT, NODE_WIDTH} from '../../constants/nodeSizes';
import {
  LINEAGE_PATH,
  PROJECT_REPOS_PATH,
  PROJECT_PIPELINES_PATH,
  TUTORIAL_PATH,
} from '../../constants/projectPaths';
import ProjectTutorial from '../../tutorials/ProjectTutorial';

import {useProjectView} from './hooks/useProjectView';
import styles from './ProjectDetails.module.css';

const ProjectDetails: React.FC = () => {
  const {repoId, pipelineId} = useUrlState();

  const {
    dags,
    nodes,
    error,
    loading,
    isSidebarOpen,
    sidebarSize,
    repoRedirect,
    pipelineRedirect,
  } = useProjectView(NODE_WIDTH, NODE_HEIGHT);

  return (
    <>
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
        <>
          <Route path={PROJECT_REPOS_PATH}>
            <ListView
              items={nodes.repos}
              getNodePath={repoRedirect}
              selectedItem={repoId}
            />
          </Route>
          <Route path={PROJECT_PIPELINES_PATH}>
            <ListView
              items={nodes.pipelines}
              getNodePath={pipelineRedirect}
              selectedItem={pipelineId}
            />
          </Route>
          <Route path={LINEAGE_PATH}>
            <DAGView dags={dags} loading={loading} error={error} />
          </Route>
        </>
      )}
      <Route path={TUTORIAL_PATH}>
        <ProjectSidebar />
        <ProjectTutorial />
      </Route>
    </>
  );
};

export default ProjectDetails;
