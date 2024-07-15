import {rest} from 'msw';

import {ClusterInfo, InspectClusterRequest} from '@dash-frontend/api/admin';
import {Empty} from '@dash-frontend/api/googleTypes';

const CLUSTER: ClusterInfo = {
  id: 'fc50aef87f1c420d80719dd649e163a2',
  deploymentId: '9pMVFEz6pjiKrqbmTPgXqpoqPybE0KJj',
  warningsOk: true,
  proxyHost: '127.0.0.1',
  proxyTls: false,
  paused: false,
  webResources: {
    archiveDownloadBaseUrl: 'http://127.0.0.1/archive/',
    createPipelineRequestJsonSchemaUrl:
      'http://127.0.0.1/jsonschema/pps_v2/CreatePipelineRequest.schema.json',
  },
  metadata: {
    clusterMetadataKey: 'clusterMetadataValue',
  },
};

export const mockInspectCluster = () =>
  rest.post<InspectClusterRequest, Empty, ClusterInfo>(
    '/api/admin_v2.API/InspectCluster',
    (_req, res, ctx) => {
      return res(ctx.json(CLUSTER));
    },
  );
