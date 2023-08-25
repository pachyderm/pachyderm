import classnames from 'classnames';
import React from 'react';

import styles from './Page.module.css';

export interface PageProps {
  children?: React.ReactNode;
  fullHeight?: boolean;
  hasDrawerPadding?: boolean;
}

const Page: React.FC<PageProps> = ({
  children,
  fullHeight = false,
  hasDrawerPadding = false,
}) => {
  return (
    <>
      <div
        className={classnames(
          styles.base,
          {[styles.fullHeight]: fullHeight},
          {[styles.hasDrawerPadding]: hasDrawerPadding},
        )}
      >
        {children}
      </div>
    </>
  );
};

export default Page;
