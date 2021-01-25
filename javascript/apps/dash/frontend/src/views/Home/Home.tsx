import React from 'react';

import {useDAGData} from 'hooks/useDAGData';

import DAG from './components/DAG';
import styles from './Home.module.css';

const Home: React.FC = () => {
  const {dag, loading, error} = useDAGData();

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

export default Home;
