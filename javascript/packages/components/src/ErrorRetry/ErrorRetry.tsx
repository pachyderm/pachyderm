import noop from 'lodash/noop';
import React from 'react';

import ButtonLink from 'ButtonLink';
import GenericError from 'GenericError';

import styles from './ErrorRetry.module.css';

interface ErrorRetryProps {
  retry?: () => void;
}

const ErrorRetry: React.FC<ErrorRetryProps> = ({
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
