import classnames from 'classnames';
import React from 'react';
import {Route} from 'react-router';

import useSidebarInfo from '@dash-frontend/hooks/useSidebarInfo';
import {LoadingDots} from '@pachyderm/components';

import DAGView from '../../components/DAGView';
import ProjectSidebar from '../../components/ProjectSidebar';
import {LINEAGE_PATH} from '../../constants/projectPaths';

import styles from './DAG.module.css';
import {useDAG} from './hooks/useDAG';

const DAG: React.FC = () => {
  const {dags, error, loading, pipelineOutputsMap} = useDAG();
  const {isOpen} = useSidebarInfo();

  return (
    <>
      {loading ? (
        <div
          className={classnames(styles.loadingContainer, {
            [styles.isSidebarOpen]: isOpen,
          })}
        >
          <LoadingDots />
        </div>
      ) : (
        <>
          <Route path={LINEAGE_PATH}>
            <DAGView dags={dags} loading={loading} error={error} />
            <ProjectSidebar pipelineOutputsMap={pipelineOutputsMap} />
          </Route>
        </>
      )}
    </>
  );
};

export default DAG;
