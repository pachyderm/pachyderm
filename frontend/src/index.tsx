import * as Sentry from '@sentry/react';
import React from 'react';
import {render} from 'react-dom';
import {load as loadRudderstack} from 'rudder-sdk-js';

import '@pachyderm/polyfills';
import '@pachyderm/components/dist/style.css';

import 'styles/index.css';

import DashUI from './DashUI';
import load from './devtools/load';

process.env.NODE_ENV === 'test' && load();

Sentry.init({
  dsn: process.env.REACT_APP_SENTRY_DSN,
  environment: process.env.REACT_APP_BUILD_ENV,
  release: process.env.REACT_APP_RELEASE_VERSION,
  tracesSampleRate: 1.0,
});

if (process.env.REACT_APP_RUDDERSTACK_ID) {
  loadRudderstack(
    process.env.REACT_APP_RUDDERSTACK_ID,
    'https://pachyderm-dataplane.rudderstack.com',
  );
}

render(<DashUI />, document.getElementById('root'));
