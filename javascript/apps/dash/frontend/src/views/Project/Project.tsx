import React from 'react';
import {useParams} from 'react-router';

import {useDAGData} from '@dash-frontend/hooks/useDAGData';

import DAG from './components/DAG';
import styles from './Project.module.css';

const Project: React.FC = () => {
  const {projectId} = useParams<{projectId: string}>();
  const {dag, loading, error} = useDAGData(projectId);

  if (error) return <h1 className={styles.base}>{JSON.stringify(error)}</h1>;
  if (loading || !dag) return <h1 className={styles.base}>Loading...</h1>;
  return (
    <div className={styles.base}>
      <div className={styles.leftPane}>
        <h1>Dash Home</h1>
      </div>
      <DAG data={dag} id="d3-dag" nodeWidth={149} nodeHeight={102} />
    </div>
  );
};

export default Project;
