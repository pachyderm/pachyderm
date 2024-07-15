import {enterprise} from './base/api';
import {getRequestOptions} from './utils/requestHeaders';

export const getEnterpriseState = () => {
  return enterprise.GetState({}, getRequestOptions());
};

// Export all of the enterprise types
export * from '@dash-frontend/generated/proto/enterprise/enterprise.pb';
