import React from 'react';

import styles from './EmptyState.module.css';

type EmptyStateProps = {
  title: string;
  message?: string;
  connect?: boolean;
  className?: string;
};

const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  message = null,
  children = null,
  className,
}) => {
  return (
    <div className={`${styles.base} ${className}`}>
      <img
        src="/elephant_empty_state.png"
        className={styles.elephantImage}
        alt=""
      />
      <h6 className={styles.title}>{title}</h6>
      <span className={styles.message}>
        {message}
        {children}
      </span>
    </div>
  );
};

export default EmptyState;
