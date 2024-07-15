import {version} from './base/api';
import {getRequestOptions} from './utils/requestHeaders';

export const getVersion = () => {
  return version.GetVersion({}, getRequestOptions());
};

// Export all of the version types
export * from '@dash-frontend/generated/proto/version/versionpb/version.pb';
