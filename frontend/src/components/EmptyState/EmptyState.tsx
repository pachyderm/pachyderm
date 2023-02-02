import React, {ReactNode} from 'react';

import {
  Link,
  ElephantEmptyState,
  ElephantErrorState,
  ExternalLinkSVG,
  Icon,
} from '@pachyderm/components';

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
      {error ? (
        <ElephantErrorState className={styles.elephantImage} />
      ) : (
        <ElephantEmptyState className={styles.elephantImage} />
      )}
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
