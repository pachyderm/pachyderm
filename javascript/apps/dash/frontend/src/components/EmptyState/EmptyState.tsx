import {ElephantEmptyState} from '@pachyderm/components';
import React from 'react';

import styles from './EmptyState.module.css';

type EmptyStateProps = {
  title: string;
  message: string;
};

const EmptyState: React.FC<EmptyStateProps> = ({title, message}) => {
  return (
    <div className={styles.base}>
      <ElephantEmptyState className={styles.elephantSvg} />
      <span className={styles.title}>{title}</span>
      <span className={styles.message}>{message}</span>
    </div>
  );
};

export default EmptyState;
