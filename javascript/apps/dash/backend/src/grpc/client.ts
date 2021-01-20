import createCredentials from './createCredentials';
import pfs from './services/pfs';
import pps from './services/pps';

const client = (address: string, authToken: string) => {
  const credentials = createCredentials(authToken);

  let pfsClient: ReturnType<typeof pfs> | undefined;
  let ppsClient: ReturnType<typeof pps> | undefined;

  // NOTE: These service clients are singletons, as we
  // don't want to create a new instance of APIClient for
  // every call stream in a transaction.
  return {
    pfs: () => {
      if (pfsClient) return pfsClient;

      pfsClient = pfs(address, credentials);
      return pfsClient;
    },
    pps: () => {
      if (ppsClient) return ppsClient;

      ppsClient = pps(address, credentials);
      return ppsClient;
    },
  };
};

export default client;
