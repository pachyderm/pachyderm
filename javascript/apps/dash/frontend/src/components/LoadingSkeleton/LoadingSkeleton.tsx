import React from 'react';
import {useRouteMatch} from 'react-router';

import {PROJECT_PATH} from '@dash-frontend/views/Project/constants/projectPaths';

import styles from './LoadingSkeleton.module.css';

const LoadingSkeleton: React.FC = () => {
  const projectMatch = useRouteMatch({
    path: PROJECT_PATH,
    exact: true,
  });

  return (
    <>
      <div className={styles.loadingHeader} />
      <div className={styles.loadingBase}>
        <div className={styles.projects} />
        {!projectMatch && <div className={styles.sidebar} />}
      </div>
    </>
  );
};

export default LoadingSkeleton;
