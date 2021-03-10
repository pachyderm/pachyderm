import classnames from 'classnames';
import React from 'react';
import {Helmet} from 'react-helmet';

import {PageHeading} from './components/PageHeading';
import styles from './Page.module.css';

export interface PageProps {
  fullHeight?: boolean;
  hasDrawerPadding?: boolean;
  title?: string;
}

const Page: React.FC<PageProps> = ({
  children,
  title,
  fullHeight = false,
  hasDrawerPadding = false,
}) => {
  return (
    <>
      {title && (
        <Helmet>
          <title>{`${title} - Pachyderm Hub`}</title>
        </Helmet>
      )}
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

export default Object.assign(Page, {Heading: PageHeading});
