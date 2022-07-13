import {ApolloError} from '@apollo/client';
import {LoadingDots} from '@pachyderm/components';
import classnames from 'classnames';
import React from 'react';
import {Route, Redirect} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Dag, DagNodes} from '@dash-frontend/lib/types';
import {lineageRoute} from '@dash-frontend/views/Project/utils/routes';

import DAGView from '../../components/DAGView';
import ListView from '../../components/ListView';
import ProjectSidebar from '../../components/ProjectSidebar';
import {
  PROJECT_PATH,
  LINEAGE_PATH,
  LINEAGE_REPOS_PATH,
  LINEAGE_PIPELINES_PATH,
  PROJECT_REPOS_PATH,
  PROJECT_PIPELINES_PATH,
} from '../../constants/projectPaths';

import {useProjectDetails} from './hooks/useProjectDetails';
import styles from './ProjectDetails.module.css';

type ProjectDetailsProps = {
  dags: Dag[] | undefined;
  nodes: {
    repos: DagNodes[];
    pipelines: DagNodes[];
  };
  error: ApolloError | undefined;
  loading: boolean;
  inputRepoLinks: Record<string, string[]>;
};

const Wrapper: React.FC = ({children}) => (
  <>
    <Route path={PROJECT_PATH}>
      <div className={styles.listWrapper}>{children}</div>
    </Route>
    <Route path={LINEAGE_PATH}>{children}</Route>
  </>
);

const ProjectDetails: React.FC<ProjectDetailsProps> = ({
  dags,
  nodes,
  error,
  loading,
  inputRepoLinks,
}) => {
  const {repoId, pipelineId, projectId} = useUrlState();
  const {isSidebarOpen, sidebarSize, repoRedirect, pipelineRedirect} =
    useProjectDetails();

  return (
    <Wrapper>
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
            <Route path={[LINEAGE_REPOS_PATH, LINEAGE_PIPELINES_PATH]} exact>
              <Redirect
                to={lineageRoute({
                  projectId,
                })}
              />
            </Route>
            <DAGView dags={dags} loading={loading} error={error} />
            <ProjectSidebar />
          </Route>
        </>
      )}
      <Route path={[PROJECT_REPOS_PATH, PROJECT_PIPELINES_PATH]}>
        <ProjectSidebar
          resizable={false}
          dagsLoading={loading}
          dagLinks={inputRepoLinks}
        />
      </Route>
    </Wrapper>
  );
};

export default ProjectDetails;
