import React from 'react';
import {Route} from 'react-router';

import DAG from './components/DAG';
import ProjectSidebar from './components/ProjectSidebar';
import {useProjectView} from './hooks/useProjectView';
import styles from './Project.module.css';

const Project: React.FC = () => {
  const {dagCount, dags, error, loading, path} = useProjectView();

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
            nodeWidth={120}
            nodeHeight={60}
            count={dagCount}
          />
        );
      })}

      <Route path={`${path}/jobs`} component={ProjectSidebar} />
    </div>
  );
};

export default Project;
