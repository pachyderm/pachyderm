import classnames from 'classnames';
import React, {ReactNode} from 'react';

import {Link, ExternalLinkSVG, Icon} from '@pachyderm/components';

import BrandedDocLink from '../BrandedDocLink';
import {BrandedEmptyIcon, BrandedErrorIcon} from '../BrandedIcon';

import styles from './EmptyState.module.css';

type EmptyStateProps = {
  children?: React.ReactNode;
  title: string | ReactNode;
  message?: string | ReactNode;
  connect?: boolean;
  className?: string;
  error?: boolean;
  renderButton?: ReactNode;
  linkToDocs?: {text: string; pathWithoutDomain: string};
  linkExternal?: {text: string; link: string};
};

const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  message = null,
  children = null,
  linkToDocs,
  linkExternal,
  className,
  error,
  renderButton,
}) => {
  return (
    <div className={classnames(styles.base, className)}>
      {error ? <BrandedErrorIcon /> : <BrandedEmptyIcon />}
      {title && (
        <span className={styles.title}>
          <h6>{title}</h6>
        </span>
      )}
      <div className={styles.createRepoButton}>{renderButton}</div>
      <span className={styles.message}>
        {message}
        {children}
      </span>
      {linkToDocs && (
        <BrandedDocLink
          externalLink
          pathWithoutDomain={linkToDocs.pathWithoutDomain}
          className={styles.docsLink}
        >
          {linkToDocs.text}
        </BrandedDocLink>
      )}
      {linkExternal && (
        <Link externalLink to={linkExternal.link} className={styles.docsLink}>
          {linkExternal.text}
          <Icon small color="inherit">
            <ExternalLinkSVG aria-hidden />
          </Icon>
        </Link>
      )}
    </div>
  );
};

export default EmptyState;
