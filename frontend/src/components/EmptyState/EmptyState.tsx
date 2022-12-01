import React from 'react';

import {ElephantEmptyState, ElephantErrorState} from '@pachyderm/components';

import styles from './EmptyState.module.css';

type EmptyStateProps = {
  title: string;
  message?: string;
  connect?: boolean;
  className?: string;
  error?: boolean;
};

const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  message = null,
  children = null,
  className,
  error,
}) => {
  return (
    <div className={`${styles.base} ${className}`}>
      {error ? (
        <ElephantErrorState className={styles.elephantImage} />
      ) : (
        <ElephantEmptyState className={styles.elephantImage} />
      )}
      <span className={styles.title}>
        <h6>{title}</h6>
      </span>
      <span className={styles.message}>
        {message}
        {children}
      </span>
    </div>
  );
};

export default EmptyState;
