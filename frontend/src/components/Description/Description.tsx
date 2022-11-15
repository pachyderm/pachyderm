import {ApolloError} from '@apollo/client';
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
  term: string;
  loading?: boolean;
  lines?: number;
  error?: ApolloError;
}

const Description: React.FC<DescriptionProps> = ({
  term,
  children,
  loading = false,
  lines = 1,
  error,
  ...rest
}) => {
  return (
    <>
      <dt className={styles.term}>
        <CaptionTextSmall>{term}</CaptionTextSmall>
      </dt>
      <dd className={styles.description} {...rest}>
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
