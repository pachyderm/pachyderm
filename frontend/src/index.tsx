import * as Sentry from '@sentry/react';
import React from 'react';
import {createRoot} from 'react-dom/client';
import {load as loadRudderstack} from 'rudder-sdk-js';

import '@pachyderm/components/index.module.css';

import 'styles/index.css';

import DashUI from './DashUI';
import load from './devtools/load';
import {getDisableTelemetry} from './lib/runtimeVariables';

const enableTelemetry = !getDisableTelemetry();

process.env.NODE_ENV === 'test' && load();

if (enableTelemetry && process.env.REACT_APP_SENTRY_DSN) {
  Sentry.init({
    dsn: process.env.REACT_APP_SENTRY_DSN,
    enabled: enableTelemetry,
    environment: process.env.REACT_APP_BUILD_ENV,
    release: process.env.REACT_APP_RELEASE_VERSION,
    tracesSampleRate: 1.0,
  });
}

if (enableTelemetry && process.env.REACT_APP_RUDDERSTACK_ID) {
  loadRudderstack(
    process.env.REACT_APP_RUDDERSTACK_ID,
    'https://pachyderm-dataplane.rudderstack.com',
  );
}

const container = document.getElementById('root');

if (container) {
  createRoot(container).render(<DashUI />);
}
