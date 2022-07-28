import {Icon, StatusWarningSVG} from '@pachyderm/components';
import React from 'react';

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
      <img
        src={error ? '/elephant_error_state.svg' : '/elephant_empty_state.png'}
        className={styles.elephantImage}
        alt=""
      />
      <span className={styles.title}>
        {error && (
          <Icon small className={styles.errorStatus}>
            <StatusWarningSVG />
          </Icon>
        )}
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
