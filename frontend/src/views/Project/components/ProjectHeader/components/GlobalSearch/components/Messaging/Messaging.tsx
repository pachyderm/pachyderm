import React from 'react';

import {PlaceholderText} from '@pachyderm/components';

import styles from './Messaging.module.css';

export const SectionHeader = ({children}: {children?: React.ReactNode}) => {
  return <span className={styles.text}>{children}</span>;
};

export const NotFoundMessage = ({children}: {children?: React.ReactNode}) => {
  return (
    <div className={styles.messagingWrapper}>
      <PlaceholderText>{children}</PlaceholderText>
    </div>
  );
};
