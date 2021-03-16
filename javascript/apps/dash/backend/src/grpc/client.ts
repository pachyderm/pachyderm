import {Metadata} from '@grpc/grpc-js';
import Logger from 'bunyan';

import createCredentials from './createCredentials';
import auth from './services/auth';
import pfs from './services/pfs';
import pps from './services/pps';
import projects from './services/projects';

interface ClientArgs {
  pachdAddress?: string;
  authToken?: string;
  projectId?: string;
  log: Logger;
}

const client = ({
  pachdAddress = '',
  authToken = '',
  projectId = '',
  log: baseLogger,
}: ClientArgs) => {
  const channelCredentials = createCredentials(pachdAddress);

  const credentialMetadata = new Metadata();
  credentialMetadata.add('authn-token', authToken);
  credentialMetadata.add('project-id', projectId);

  const log = baseLogger.child({
    eventSource: 'grpc client',
    pachdAddress,
    projectId,
  });

  let pfsClient: ReturnType<typeof pfs> | undefined;
  let ppsClient: ReturnType<typeof pps> | undefined;
  let authClient: ReturnType<typeof auth> | undefined;
  let projectsClient: ReturnType<typeof projects> | undefined;

  // NOTE: These service clients are singletons, as we
  // don't want to create a new instance of APIClient for
  // every call stream in a transaction.
  return {
    pfs: () => {
      if (pfsClient) return pfsClient;

      pfsClient = pfs({
        pachdAddress,
        channelCredentials,
        credentialMetadata,
        log,
      });
      return pfsClient;
    },
    pps: () => {
      if (ppsClient) return ppsClient;

      ppsClient = pps({
        pachdAddress,
        channelCredentials,
        credentialMetadata,
        log,
      });
      return ppsClient;
    },
    auth: () => {
      if (authClient) return authClient;

      authClient = auth({
        pachdAddress,
        channelCredentials,
        credentialMetadata,
        log,
      });
      return authClient;
    },
    projects: () => {
      if (projectsClient) return projectsClient;

      projectsClient = projects({
        pachdAddress,
        channelCredentials,
        credentialMetadata,
        log,
      });
      return projectsClient;
    },
  };
};

export default client;
