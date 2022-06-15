import {ApolloError} from '@apollo/client';
import {
  Button,
  StatusWarningSVG,
  GenericErrorSVG,
  Group,
  ErrorText,
} from '@pachyderm/components';
import {GraphQLError} from 'graphql';
import React from 'react';
import {Helmet} from 'react-helmet';

import JSONDataPreview from '@dash-frontend/components/JSONDataPreview';
import View from '@dash-frontend/components/View';
import LandingHeader from '@dash-frontend/views/Landing/components/LandingHeader';

import styles from './ErrorView.module.css';
import useErrorView, {ErrorViewType} from './hooks/useErrorView';

type ErrorViewProps = {
  apolloError?: ApolloError;
  graphQLError?: GraphQLError;
};

const ErrorView: React.FC<ErrorViewProps> = ({graphQLError}) => {
  const {errorType, errorMessage} = useErrorView(graphQLError);

  return (
    <>
      <Helmet>
        <title>Error - Pachyderm Console</title>
      </Helmet>
      <LandingHeader />
      <View>
        <Group vertical spacing={16} align="center" className={styles.base}>
          <GenericErrorSVG />
          <Group vertical spacing={32} className={styles.content}>
            <Group spacing={8} align="center">
              <StatusWarningSVG /> <h4>{errorMessage}</h4>
            </Group>
            <ErrorText>
              {graphQLError?.extensions?.exception?.stacktrace
                ? graphQLError?.extensions?.exception?.stacktrace[0]
                : String(graphQLError?.extensions?.details)}
            </ErrorText>
            {graphQLError?.source && (
              <>
                Source: <ErrorText>{graphQLError?.source}</ErrorText>
              </>
            )}
            {errorType !== ErrorViewType.UNAUTHENTICATED && (
              <Button href="/">Go Back Home</Button>
            )}
          </Group>

          {graphQLError && (
            <div className={styles.fullError}>
              <JSONDataPreview
                inputData={graphQLError}
                formattingStyle="json"
                width={Math.min(window.innerWidth - 128, 948)}
              />
            </div>
          )}
        </Group>
      </View>
    </>
  );
};

export default ErrorView;
