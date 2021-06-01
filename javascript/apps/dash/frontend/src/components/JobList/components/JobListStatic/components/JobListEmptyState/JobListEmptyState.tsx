import {ElephantEmptyState} from '@pachyderm/components';
import React from 'react';

import styles from './JobListEmptyState.module.css';

type JobListEmptyStateProps = {
  title: string;
  message: string;
};

const JobListEmptyState: React.FC<JobListEmptyStateProps> = ({
  title,
  message,
}) => {
  return (
    <div className={styles.base}>
      <ElephantEmptyState className={styles.elephantSvg} />
      <span className={styles.title}>{title}</span>
      <span className={styles.message}>{message}</span>
    </div>
  );
};

export default JobListEmptyState;
