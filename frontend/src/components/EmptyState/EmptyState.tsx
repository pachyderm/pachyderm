import React, {ReactNode} from 'react';

import {Link, ExternalLinkSVG, Icon} from '@pachyderm/components';

import {BrandedEmptyIcon, BrandedErrorIcon} from '../BrandedIcon';

import styles from './EmptyState.module.css';

type EmptyStateProps = {
  title: string;
  message?: string;
  connect?: boolean;
  className?: string;
  error?: boolean;
  renderButton?: ReactNode;
  linkToDocs?: {text: string; link: string};
};

const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  message = null,
  children = null,
  linkToDocs,
  className,
  error,
  renderButton,
}) => {
  return (
    <div className={`${styles.base} ${className}`}>
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
        <Link externalLink to={linkToDocs.link} className={styles.docsLink}>
          {linkToDocs.text}
          <Icon small>
            <ExternalLinkSVG aria-hidden />
          </Icon>
        </Link>
      )}
    </div>
  );
};

export default EmptyState;
