const defaultValue = 10000;
const minValue = 100;

const getPollingIntervalFromEnv = () => {
  const envVar = process.env.REACT_APP_POLLING;
  if (!envVar) {
    return defaultValue;
  }

  const value = +envVar;

  if (value === 0) {
    return value;
  }

  if (!+value) {
    console.error(
      `environment variable REACT_APP_POLLING appears to not be a number. Using a default value of ${defaultValue} instead.`,
    );
    return defaultValue;
  }

  if (value <= minValue) {
    console.error(
      `environment variable REACT_APP_POLLING was set to a number below ${minValue}ms. Using a value of ${minValue}ms instead.`,
    );
    return minValue;
  }
  return value;
};

const pollingInterval = getPollingIntervalFromEnv();

export const JOBS_POLL_INTERVAL_MS: number = pollingInterval;
export const PROJECTS_POLL_INTERVAL_MS: number = pollingInterval;
export const REPO_POLL_INTERVAL_MS: number = pollingInterval;
export const PIPELINES_POLL_INTERVAL_MS: number = pollingInterval;
export const COMMITS_POLL_INTERVAL_MS: number = pollingInterval;
export const LOGS_POLL_INTERVAL_MS: number = pollingInterval;
export const CLUSTER_DEFAULTS_POLL_INTERVAL_MS: number = pollingInterval;
export const DEFAULT_POLLING_INTERVAL_MS = pollingInterval;
