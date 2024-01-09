import React from 'react';
import {Helmet} from 'react-helmet';

import CodePreview from '@dash-frontend/components/CodePreview';
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
  stackTrace?: string | object | unknown[];
  showBackHomeButton?: boolean;
};

const ErrorView: React.FC<ErrorViewProps> = ({
  showBackHomeButton = false,
  errorMessage,
  errorDetails,
  source,
  stackTrace,
}) => {
  const stackTraceIsString = typeof stackTrace === 'string';

  return (
    <>
      <Helmet>
        <title>{'Console - Error'}</title>
      </Helmet>
      <LandingHeader disableBranding />
      <View>
        <Group vertical spacing={16} align="center" className={styles.base}>
          <Group vertical spacing={32} className={styles.content}>
            <Group spacing={8} align="center">
              <Icon color="red">
                <StatusWarningSVG />
              </Icon>{' '}
              <h4>{errorMessage}</h4>
            </Group>
            {errorDetails && <ErrorText>{errorDetails}</ErrorText>}
            {source && <ErrorText>Source: {source}</ErrorText>}
            {showBackHomeButton && (
              <Button tabIndex={-1}>
                <a href="/" className={styles.backLink}>
                  Go Back Home
                </a>
              </Button>
            )}
          </Group>

          {stackTrace && (
            <CodePreview
              className={styles.fullError}
              source={
                stackTraceIsString
                  ? stackTrace
                  : JSON.stringify(stackTrace, null, 2)
              }
              language={stackTraceIsString ? 'text' : 'json'}
              hideGutter
            />
          )}
        </Group>
      </View>
    </>
  );
};

export default ErrorView;
