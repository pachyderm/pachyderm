import React from 'react';
import {Switch, Route} from 'react-router';

import styles from './LoadingSkeleton.module.css';

const LoadingSkeleton: React.FC = () => {
  return (
    <>
      <div className={styles.loadingHeader} />
      <div className={styles.loadingBase}>
        <div className={styles.projects} />
        <Switch>
          <Route path="/" exact>
            <div className={styles.sidebar} />
          </Route>
        </Switch>
      </div>
    </>
  );
};

export default LoadingSkeleton;
