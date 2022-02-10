export const getIssuerUri = () => {
  // when running in the node runtime environment,
  // environment variables can't be objects. Therefore
  // we'll need jest to override this in order to be set.
  let issuerUri = '';
  if (window.pachDashConfig) {
    issuerUri = window.pachDashConfig.REACT_APP_RUNTIME_ISSUER_URI;
  } else if (process.env.REACT_APP_RUNTIME_ISSUER_URI) {
    issuerUri = process.env.REACT_APP_RUNTIME_ISSUER_URI;
  } else {
    issuerUri = process.env.pachDashConfig.REACT_APP_RUNTIME_ISSUER_URI || '';
  }

  return issuerUri;
};

export const getSubscriptionsPrefix = () => {
  let subscriptionsPrefix = '';
  if (window.pachDashConfig) {
    subscriptionsPrefix =
      window.pachDashConfig.REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX;
  } else if (process.env.REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX) {
    subscriptionsPrefix = process.env.REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX;
  } else {
    subscriptionsPrefix =
      process.env.pachDashConfig.REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX || '';
  }

  return subscriptionsPrefix;
};
