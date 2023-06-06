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

export const getProxyHostName = () => {
  let proxyHost = '';
  if (window.pachDashConfig) {
    proxyHost =
      window.pachDashConfig.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_HOST || '';
  } else {
    proxyHost = process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_HOST || '';
  }

  return proxyHost;
};

export const getProxyEnabled = () => {
  return getProxyHostName() !== '';
};

export const getTlsEnabled = () => {
  let tlsEnabled = false;
  if (window.pachDashConfig) {
    tlsEnabled =
      window.pachDashConfig.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_TLS === 'true';
  } else {
    tlsEnabled = process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_TLS === 'true';
  }

  return tlsEnabled;
};
