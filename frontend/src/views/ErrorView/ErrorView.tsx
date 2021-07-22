import {
  Button,
  ExclamationErrorSVG,
  GenericError,
  Group,
} from '@pachyderm/components';
import React from 'react';

import View from '@dash-frontend/components/View';

import styles from './ErrorView.module.css';
import useErrorView, {ErrorViewType} from './hooks/useErrorView';

const ErrorView = () => {
  const {errorType} = useErrorView();

  return (
    <View>
      <GenericError>
        <h1 className={styles.heading}>
          {errorType === ErrorViewType.NOT_FOUND ? (
            <>Elephants never forget, so this page must not exist.</>
          ) : (
            <Group spacing={8} align="center">
              <ExclamationErrorSVG className={styles.error} />{' '}
              <span>
                {errorType === ErrorViewType.UNAUTHENTICATED
                  ? 'Unable to authenticate. Try again later.'
                  : 'Something went wrong. Try again Later.'}
              </span>
            </Group>
          )}
        </h1>

        <Button href="/" autoWidth className={styles.homeButton}>
          Go Back Home
        </Button>
      </GenericError>
    </View>
  );
};

export default ErrorView;
