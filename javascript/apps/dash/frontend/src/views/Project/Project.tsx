import React from 'react';

import DAG from './components/DAG';
import ProjectSidebar from './components/ProjectSidebar';
import {NODE_HEIGHT, NODE_WIDTH} from './constants/nodeSizes';
import {useProjectView} from './hooks/useProjectView';
import styles from './Project.module.css';

const Project: React.FC = () => {
  const {dagCount, dags, error, loading} = useProjectView(
    NODE_WIDTH,
    NODE_HEIGHT,
  );

  if (error) return <h1 className={styles.base}>{JSON.stringify(error)}</h1>;
  if (loading || !dags) return <h1 className={styles.base}>Loading...</h1>;
  return (
    <div className={styles.wrapper}>
      {dags.map((dag, i) => {
        return (
          <DAG
            data={dag}
            key={i}
            id={`dag${i}`}
            nodeWidth={NODE_WIDTH}
            nodeHeight={NODE_HEIGHT}
            count={dagCount}
          />
        );
      })}

      <ProjectSidebar />
    </div>
  );
};

export default Project;
