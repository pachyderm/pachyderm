import {Metadata} from '@grpc/grpc-js';

import {GRPCPlugin, ServiceDefinition} from './lib/types';
import authServiceRpcHandler from './services/auth';
import pfsServiceRpcHandler from './services/pfs';

interface ClientArgs {
  authToken?: string;
  plugins?: GRPCPlugin[];
}

const attachPlugins = <T extends ServiceDefinition>(
  service: T,
  plugins: GRPCPlugin[] = [],
): T => {
  const onCallObservers = plugins.flatMap((p) => (p.onCall ? [p.onCall] : []));
  const onCompleteObservers = plugins.flatMap((p) =>
    p.onCompleted ? [p.onCompleted] : [],
  );
  const onErrorObservers = plugins.flatMap((p) =>
    p.onError ? [p.onError] : [],
  );

  const serviceProxyHandler: ProxyHandler<T> = {
    // TS doesn't support symbol indexing
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    get: (service, requestName: any) => {
      // const requestName = String(key);
      // technically, a key can be a symbol
      const originalHandler = service[requestName];

      return async (...args: never[]) => {
        try {
          onCallObservers.forEach((cb) => cb({requestName}));
          const result = await originalHandler(...args);
          onCompleteObservers.forEach((cb) => cb({requestName}));
          return result;
        } catch (e) {
          onErrorObservers.forEach((cb) => cb({error: e, requestName}));
          throw e;
        }
      };
    },
  };

  return new Proxy(service, serviceProxyHandler);
};

function bindPluginsToServices<T extends ServiceDefinition>(
  services: Record<string, T>,
  plugins: GRPCPlugin[],
) {
  for (const [key, value] of Object.entries(services)) {
    services[key] = attachPlugins(value, plugins);
  }
}

/**
 * The apiClientRequestWrapper function performs the following tasks:
 * - It establishes a GRPC API client for every service in Pachyderm, injecting
 *     a requests metadata like the authentication token into RPC service calls.
 * - It loads the plugins' proxies to the services.
 * - Lastly, it returns these services encapsulated in an object.
 */
const apiClientRequestWrapper = ({
  authToken = '',
  plugins = [],
}: ClientArgs = {}) => {
  const credentialMetadata = new Metadata();
  credentialMetadata.add('authn-token', authToken);
  credentialMetadata.add('project-id', '');

  const enrichedServices = {
    pfs: pfsServiceRpcHandler({credentialMetadata, plugins}),
    auth: authServiceRpcHandler({credentialMetadata}),
  };
  bindPluginsToServices(enrichedServices, plugins);

  return {
    ...enrichedServices,
    attachCredentials: (header: string, value: string) => {
      credentialMetadata.set(header, value);
    },
  };
};

export default apiClientRequestWrapper;
