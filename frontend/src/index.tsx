import * as Sentry from '@sentry/react';
import LogRocket from 'logrocket';
import React from 'react';
import {render} from 'react-dom';
import {load as loadRudderstack} from 'rudder-sdk-js';

import '@pachyderm/components/dist/style.css';

import 'styles/index.css';

import DashUI from './DashUI';
import load from './devtools/load';
import {getDisableTelemetry} from './lib/runtimeVariables';

const enableTelemetry = !getDisableTelemetry();

process.env.NODE_ENV === 'test' && load();

Sentry.init({
  dsn: process.env.REACT_APP_SENTRY_DSN,
  enabled: enableTelemetry,
  environment: process.env.REACT_APP_BUILD_ENV,
  release: process.env.REACT_APP_RELEASE_VERSION,
  tracesSampleRate: 1.0,
  beforeSend(event) {
    if (
      event?.extra &&
      process.env.REACT_APP_LOGROCKET_ID &&
      LogRocket.sessionURL !== null
    ) {
      event.extra['LogRocketError'] = LogRocket.sessionURL;
    }
    return event;
  },
});

if (enableTelemetry && process.env.REACT_APP_RUDDERSTACK_ID) {
  loadRudderstack(
    process.env.REACT_APP_RUDDERSTACK_ID,
    'https://pachyderm-dataplane.rudderstack.com',
  );
}

if (enableTelemetry && process.env.REACT_APP_LOGROCKET_ID) {
  LogRocket.init(process.env.REACT_APP_LOGROCKET_ID);

  LogRocket.getSessionURL((sessionURL) => {
    Sentry.configureScope((scope) => {
      scope.setExtra('LogRocketSession', sessionURL);
    });
  });
}

render(<DashUI />, document.getElementById('root'));
