import classNames from 'classnames';
import React from 'react';

import {PlaceholderText} from '@pachyderm/components';

import styles from './Messaging.module.css';

export const NotFoundMessage = ({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) => {
  return (
    <div className={classNames(styles.messagingWrapper, className)}>
      <PlaceholderText>{children}</PlaceholderText>
    </div>
  );
};
