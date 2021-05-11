import React from 'react';

import styles from './Messaging.module.css';

export const SectionHeader: React.FC = ({children}) => {
  return <span className={styles.text}>{children}</span>;
};

export const NotFoundMessage: React.FC = ({children}) => {
  return (
    <div className={styles.messagingWrapper}>
      <span className={styles.text}>{children}</span>
    </div>
  );
};
