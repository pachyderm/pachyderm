import React from 'react';

import styles from './LoadingDots.module.css';

const LoadingDots: React.FC = () => {
  return (
    <div className={styles.container}>
      <div className={styles.base} role="status" aria-label="loading" />
    </div>
  );
};

export default LoadingDots;
