import {EditMetadataRequest} from '@dash-frontend/generated/proto/metadata/metadata.pb';

import {metadata} from './base/api';
import {getRequestOptions} from './utils/requestHeaders';

export const editMetadata = (req: EditMetadataRequest = {}) => {
  return metadata.EditMetadata(req, getRequestOptions());
};

// Export all of the metadata types
export * from '@dash-frontend/generated/proto/metadata/metadata.pb';
