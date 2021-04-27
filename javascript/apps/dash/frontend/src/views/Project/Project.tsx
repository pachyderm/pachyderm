import React, {useState} from 'react';
import {Redirect} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';

import DAG from './components/DAG';
import ProjectSidebar from './components/ProjectSidebar';
import {NODE_HEIGHT, NODE_WIDTH} from './constants/nodeSizes';
import {useProjectView} from './hooks/useProjectView';
import styles from './Project.module.css';
import {dagRoute} from './utils/routes';

const Project: React.FC = () => {
  const {dags, error, loading} = useProjectView(NODE_WIDTH, NODE_HEIGHT);
  const {dagId, projectId} = useUrlState();
  const [largestDagWidth, setLargestDagWidth] = useState<number | null>(null);

  if (error) return <h1 className={styles.base}>{JSON.stringify(error)}</h1>;
  if (loading || !dags) return <h1 className={styles.base}>Loading...</h1>;

  const dagFromRoute = dags.find((dag) => dag.id === dagId);
  const dagsToShow = dagFromRoute ? [dagFromRoute] : dags;

  if (!dagFromRoute && dagsToShow.length === 1) {
    return <Redirect to={dagRoute({projectId, dagId: dagsToShow[0].id})} />;
  }

  return (
    <div className={styles.wrapper}>
      {dagsToShow.map((dag) => {
        return (
          <DAG
            data={dag}
            key={dag.id}
            id={dag.id}
            nodeWidth={NODE_WIDTH}
            nodeHeight={NODE_HEIGHT}
            count={dagsToShow.length}
            isInteractive={dagsToShow.length === 1}
            largestDagWidth={largestDagWidth}
            setLargestDagWidth={setLargestDagWidth}
          />
        );
      })}

      <ProjectSidebar />
    </div>
  );
};

export default Project;
