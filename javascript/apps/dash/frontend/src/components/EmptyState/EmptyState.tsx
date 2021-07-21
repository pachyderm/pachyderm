import React from 'react';

import styles from './EmptyState.module.css';

type EmptyStateProps = {
  title: string;
  message: string;
};

const EmptyState: React.FC<EmptyStateProps> = ({title, message}) => {
  return (
    <div className={styles.base}>
      <img
        src="/elephant_empty_state.png"
        className={styles.elephantImage}
        alt=""
      />
      <span className={styles.title}>{title}</span>
      <span className={styles.message}>{message}</span>
    </div>
  );
};

export default EmptyState;
