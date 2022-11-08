import React from 'react';
import {Helmet} from 'react-helmet';

import JSONDataPreview from '@dash-frontend/components/JSONDataPreview';
import View from '@dash-frontend/components/View';
import LandingHeader from '@dash-frontend/views/Landing/components/LandingHeader';
import {
  Button,
  StatusWarningSVG,
  Group,
  ErrorText,
  Icon,
} from '@pachyderm/components';

import styles from './ErrorView.module.css';

type ErrorViewProps = {
  errorMessage?: string;
  errorDetails?: string;
  source?: string;
  stackTrace?: unknown;
  showBackHomeButton?: boolean;
};

const ErrorView: React.FC<ErrorViewProps> = ({
  showBackHomeButton = false,
  errorMessage,
  errorDetails,
  source,
  stackTrace,
}) => {
  return (
    <>
      <Helmet>
        <title>Error - Pachyderm Console</title>
      </Helmet>
      <LandingHeader />
      <View>
        <Group vertical spacing={16} align="center" className={styles.base}>
          <img
            src="elephant_error_state.svg"
            className={styles.elephantImage}
            alt="Error Encountered"
          />
          <Group vertical spacing={32} className={styles.content}>
            <Group spacing={8} align="center">
              <Icon color="red">
                <StatusWarningSVG />
              </Icon>{' '}
              <h4>{errorMessage}</h4>
            </Group>
            {errorDetails && <ErrorText>{errorDetails}</ErrorText>}
            {source && <ErrorText>Source: {source}</ErrorText>}
            {showBackHomeButton && <Button href="/">Go Back Home</Button>}
          </Group>

          {stackTrace && (
            <div className={styles.fullError}>
              <JSONDataPreview
                inputData={stackTrace}
                formattingStyle="yaml"
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
