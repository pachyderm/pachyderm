export const getIssuerUri = () => {
  const {REACT_APP_RUNTIME_ISSUER_URI, pachDashConfig} = process.env;

  // when running in the node runtime environment,
  // environment variables can't be objects. Therefore
  // we'll need jest to override this in order to be set.
  let issuerUri = '';
  if (REACT_APP_RUNTIME_ISSUER_URI) {
    issuerUri = REACT_APP_RUNTIME_ISSUER_URI;
  } else {
    issuerUri = pachDashConfig.REACT_APP_RUNTIME_ISSUER_URI;
  }

  return issuerUri;
};

export const getSubscriptionsPrefix = () => {
  const {REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX, pachDashConfig} = process.env;

  let subscriptionsPrefix = '';
  if (REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX) {
    subscriptionsPrefix = REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX;
  } else {
    subscriptionsPrefix = pachDashConfig.REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX;
  }

  return subscriptionsPrefix;
};
