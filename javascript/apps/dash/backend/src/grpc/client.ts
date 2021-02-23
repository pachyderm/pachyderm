import {Metadata} from '@grpc/grpc-js';

import createCredentials from './createCredentials';
import auth from './services/auth';
import pfs from './services/pfs';
import pps from './services/pps';

const client = (address: string, authToken = '', projectId = '') => {
  const channelCredentials = createCredentials(address);

  const credentialMetadata = new Metadata();
  credentialMetadata.add('authn-token', authToken);
  credentialMetadata.add('project-id', projectId);

  let pfsClient: ReturnType<typeof pfs> | undefined;
  let ppsClient: ReturnType<typeof pps> | undefined;
  let authClient: ReturnType<typeof auth> | undefined;

  // NOTE: These service clients are singletons, as we
  // don't want to create a new instance of APIClient for
  // every call stream in a transaction.
  return {
    pfs: () => {
      if (pfsClient) return pfsClient;

      pfsClient = pfs(address, channelCredentials, credentialMetadata);
      return pfsClient;
    },
    pps: () => {
      if (ppsClient) return ppsClient;

      ppsClient = pps(address, channelCredentials, credentialMetadata);
      return ppsClient;
    },
    auth: () => {
      if (authClient) return authClient;

      authClient = auth(address, channelCredentials);
      return authClient;
    },
  };
};

export default client;
