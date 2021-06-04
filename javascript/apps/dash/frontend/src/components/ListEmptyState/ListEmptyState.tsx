import {ElephantEmptyState} from '@pachyderm/components';
import React from 'react';

import styles from './ListEmptyState.module.css';

type ListEmptyStateProps = {
  title: string;
  message: string;
};

const ListEmptyState: React.FC<ListEmptyStateProps> = ({title, message}) => {
  return (
    <div className={styles.base}>
      <ElephantEmptyState className={styles.elephantSvg} />
      <span className={styles.title}>{title}</span>
      <span className={styles.message}>{message}</span>
    </div>
  );
};

export default ListEmptyState;
