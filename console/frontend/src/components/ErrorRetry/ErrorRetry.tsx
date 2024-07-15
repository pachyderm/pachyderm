import noop from 'lodash/noop';
import React from 'react';

import {GenericError} from '@dash-frontend/components/GenericError';
import {ButtonLink} from '@pachyderm/components';

import styles from './ErrorRetry.module.css';

interface ErrorRetryProps {
  children?: React.ReactNode;
  retry?: () => void;
}

export const ErrorRetry: React.FC<ErrorRetryProps> = ({
  retry = noop,
  children,
  ...rest
}) => (
  <GenericError className={styles.base} {...rest}>
    <div className={styles.text}>
      <span className={styles.exclamation}>!</span> {children} Click{' '}
      <ButtonLink onClick={retry} aria-label="Retry fetching resource">
        here
      </ButtonLink>{' '}
      to retry.
    </div>
  </GenericError>
);

export default ErrorRetry;
