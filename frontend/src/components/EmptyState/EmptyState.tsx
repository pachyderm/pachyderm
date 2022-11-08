import classnames from 'classnames';
import React from 'react';

import {Icon, StatusWarningSVG} from '@pachyderm/components';

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
      {!error && (
        <img
          src="/elephant_empty_state.png"
          className={styles.elephantImage}
          alt=""
        />
      )}
      <span className={classnames(styles.title, {[styles.noImage]: error})}>
        {error && (
          <Icon small className={styles.errorStatus} color="red">
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
