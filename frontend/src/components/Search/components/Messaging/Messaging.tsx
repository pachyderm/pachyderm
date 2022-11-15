import React from 'react';

import {PlaceholderText} from '@pachyderm/components';

import styles from './Messaging.module.css';

export const SectionHeader: React.FC = ({children}) => {
  return <span className={styles.text}>{children}</span>;
};

export const NotFoundMessage: React.FC = ({children}) => {
  return (
    <div className={styles.messagingWrapper}>
      <PlaceholderText>{children}</PlaceholderText>
    </div>
  );
};
