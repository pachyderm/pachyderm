import {InspectClusterRequest} from '@dash-frontend/generated/proto/admin/admin.pb';

import {admin} from './base/api';
import {getRequestOptions} from './utils/requestHeaders';

export const inspectCluster = (req: InspectClusterRequest = {}) => {
  return admin.InspectCluster(req, getRequestOptions());
};

// Export all of the admin types
export * from '@dash-frontend/generated/proto/admin/admin.pb';
