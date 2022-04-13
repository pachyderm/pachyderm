import {
  Button,
  StatusWarningSVG,
  GenericError,
  Group,
} from '@pachyderm/components';
import React from 'react';
import {Helmet} from 'react-helmet';

import View from '@dash-frontend/components/View';

import styles from './ErrorView.module.css';
import useErrorView, {ErrorViewType} from './hooks/useErrorView';

const ErrorView = () => {
  const {errorType} = useErrorView();

  return (
    <View>
      <Helmet>
        <title>Error - Pachyderm Console</title>
      </Helmet>
      <GenericError>
        <div className={styles.heading}>
          {errorType === ErrorViewType.NOT_FOUND ? (
            <h2>Elephants never forget, so this page must not exist.</h2>
          ) : (
            <Group spacing={8} align="center">
              <StatusWarningSVG className={styles.error} />{' '}
              <h2>
                {errorType === ErrorViewType.UNAUTHENTICATED
                  ? 'Unable to authenticate. Try again later.'
                  : 'Something went wrong. Try again Later.'}
              </h2>
            </Group>
          )}
        </div>

        <Button href="/" autoWidth className={styles.homeButton}>
          Go Back Home
        </Button>
      </GenericError>
    </View>
  );
};

export default ErrorView;
