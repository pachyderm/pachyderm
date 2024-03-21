export const getIssuerUri = () => {
  // when running in the node runtime environment,
  // environment variables can't be objects. Therefore
  // we'll need jest to override this in order to be set.
  let issuerUri = '';
  if (window.pachDashConfig) {
    issuerUri = window.pachDashConfig.REACT_APP_RUNTIME_ISSUER_URI;
  } else {
    issuerUri = process.env.REACT_APP_RUNTIME_ISSUER_URI || '';
  }

  return issuerUri;
};

export const getSubscriptionsPrefix = () => {
  let subscriptionsPrefix = '';
  if (window.pachDashConfig) {
    subscriptionsPrefix =
      window.pachDashConfig.REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX || '';
  } else {
    subscriptionsPrefix =
      process.env.REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX || '';
  }

  return subscriptionsPrefix;
};

export const getDisableTelemetry = () => {
  let disableTelemetry = '';
  if (window.pachDashConfig) {
    disableTelemetry =
      window.pachDashConfig.REACT_APP_RUNTIME_DISABLE_TELEMETRY || '';
  } else {
    disableTelemetry = process.env.REACT_APP_RUNTIME_DISABLE_TELEMETRY || '';
  }

  return disableTelemetry === 'true';
};

export const getReleaseVersion = () => {
  let releaseVersion = '';
  if (window.pachDashConfig) {
    releaseVersion = window.pachDashConfig.REACT_APP_RELEASE_VERSION || '';
  } else {
    releaseVersion = process.env.REACT_APP_RELEASE_VERSION || '';
  }

  return releaseVersion;
};

export const getRefetchInterval = () => {
  let refetchInterval = '';
  if (window.pachDashConfig) {
    refetchInterval =
      window.pachDashConfig.REACT_APP_RUNTIME_REFETCH_INTERVAL || '';
  } else {
    refetchInterval = process.env.REACT_APP_RUNTIME_REFETCH_INTERVAL || '';
  }

  return refetchInterval;
};
