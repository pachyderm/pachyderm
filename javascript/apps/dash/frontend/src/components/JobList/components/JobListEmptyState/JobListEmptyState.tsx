import React from 'react';

import styles from './JobListEmptyState.module.css';

const JobListEmptyState: React.FC = () => {
  return (
    <div className={styles.base}>
      <div className={styles.circle} />
      <span className={styles.start}>Let&apos;s Start</span>
      <span className={styles.create}>
        Create your first job on this project! First, fix any crashing pipelines
        before you create a job.
      </span>
    </div>
  );
};

export default JobListEmptyState;
