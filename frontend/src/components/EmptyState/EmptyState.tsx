import classnames from 'classnames';
import React, {ReactNode} from 'react';

import {Link, ExternalLinkSVG, Icon, LockIconSVG} from '@pachyderm/components';

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
  noAccess?: boolean;
  renderButton?: ReactNode;
  linkToDocs?: {text: string; pathWithoutDomain: string};
  linkExternal?: {text: string; link: string};
};

const EmptyStateIcon: React.FC<Pick<EmptyStateProps, 'noAccess' | 'error'>> = ({
  noAccess,
  error,
}) => {
  if (noAccess) return <LockIconSVG />;
  if (error) return <BrandedErrorIcon />;
  return <BrandedEmptyIcon />;
};

const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  message = null,
  children = null,
  linkToDocs,
  linkExternal,
  className,
  error,
  noAccess,
  renderButton,
}) => {
  return (
    <div className={classnames(styles.base, className)}>
      <EmptyStateIcon error={error} noAccess={noAccess} />
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
