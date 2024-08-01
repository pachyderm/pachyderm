import {getRefetchInterval} from '@dash-frontend/lib/runtimeVariables';

const defaultValue = 3000;
const minValue = 100;

const getRefetchIntervalFromEnv = () => {
  const envVar = getRefetchInterval();

  if (!envVar) {
    return defaultValue;
  }

  if (envVar === 'test') {
    return false;
  }

  const value = +envVar;

  if (!value) {
    console.error(
      `environment variable REACT_APP_RUNTIME_REFETCH_INTERVAL appears to not be a number. Using a default value of ${defaultValue} instead.`,
    );
    return defaultValue;
  }

  if (value <= minValue) {
    console.error(
      `environment variable REACT_APP_RUNTIME_REFETCH_INTERVAL was set to a number below ${minValue}ms. Using a value of ${defaultValue}ms instead.`,
    );
    return defaultValue;
  }
  return value;
};

export const DEFAULT_POLLING_INTERVAL_MS = getRefetchIntervalFromEnv();
