import kebabCase from 'lodash/kebabCase';
import React, {HTMLAttributes} from 'react';

import {
  SkeletonBodyText,
  CaptionTextSmall,
  ErrorText,
  StatusWarningSVG,
  Icon,
} from '@pachyderm/components';

import styles from './Description.module.css';

interface DescriptionProps extends HTMLAttributes<HTMLElement> {
  term?: string;
  loading?: boolean;
  lines?: number;
  error?: string;
}

const Description: React.FC<DescriptionProps> = ({
  term,
  children,
  loading = false,
  lines = 1,
  error,
  'aria-label': ariaLabelIn,
  ...rest
}) => {
  const ariaLabel = kebabCase(term) || ariaLabelIn;

  return (
    <>
      {term && (
        <dt className={styles.term} id={ariaLabel}>
          <CaptionTextSmall>{term}</CaptionTextSmall>
        </dt>
      )}
      <dd className={styles.description} {...rest} aria-labelledby={ariaLabel}>
        {loading && (
          <div className={lines === 1 ? styles.singleLineLoading : undefined}>
            <SkeletonBodyText
              lines={lines}
              data-testid={`Description__${term}Skeleton`}
            />
          </div>
        )}{' '}
        {error && (
          <div className={styles.term}>
            <Icon small color="red">
              <StatusWarningSVG />
            </Icon>
            <ErrorText>Unable to load</ErrorText>
          </div>
        )}
        {!loading && !error && children}
      </dd>
    </>
  );
};

export default Description;
