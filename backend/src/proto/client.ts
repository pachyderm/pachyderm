import {Metadata} from '@grpc/grpc-js';

import createCredentials from './createCredentials';
import {GRPCPlugin, ServiceDefinition} from './lib/types';
import admin from './services/admin';
import auth from './services/auth';
import enterprise from './services/enterprise';
import license from './services/license';
import pfs from './services/pfs';
import pps from './services/pps';
import projects from './services/projects';

interface ClientArgs {
  pachdAddress?: string;
  authToken?: string;
  projectId?: string;
  plugins?: GRPCPlugin[];
  ssl?: boolean;
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

const client = ({
  pachdAddress = '',
  authToken = '',
  projectId = '',
  plugins = [],
  ssl = true,
}: ClientArgs) => {
  const channelCredentials = createCredentials(ssl);

  const credentialMetadata = new Metadata();
  credentialMetadata.add('authn-token', authToken);
  credentialMetadata.add('project-id', projectId);

  let pfsService: ReturnType<typeof pfs> | undefined;
  let ppsService: ReturnType<typeof pps> | undefined;
  let authService: ReturnType<typeof auth> | undefined;
  let projectsService: ReturnType<typeof projects> | undefined;
  let adminService: ReturnType<typeof admin> | undefined;
  let enterpriseService: ReturnType<typeof enterprise> | undefined;
  let licenseService: ReturnType<typeof license> | undefined;

  // NOTE: These service clients are singletons, as we
  // don't want to create a new instance of APIClient for
  // every call stream in a transaction.
  const methods = {
    license: () => {
      if (licenseService) return licenseService;

      licenseService = attachPlugins(
        license({
          pachdAddress,
          channelCredentials,
          credentialMetadata,
        }),
        plugins,
      );

      return licenseService;
    },
    enterprise: () => {
      if (enterpriseService) return enterpriseService;

      enterpriseService = attachPlugins(
        enterprise({
          pachdAddress,
          channelCredentials,
          credentialMetadata,
        }),
        plugins,
      );

      return enterpriseService;
    },
    pfs: () => {
      if (pfsService) return pfsService;

      pfsService = attachPlugins(
        pfs({
          pachdAddress,
          channelCredentials,
          credentialMetadata,
          plugins,
        }),
        plugins,
      );
      return pfsService;
    },
    pps: () => {
      if (ppsService) return ppsService;

      ppsService = attachPlugins(
        pps({
          pachdAddress,
          channelCredentials,
          credentialMetadata,
        }),
        plugins,
      );
      return ppsService;
    },
    auth: () => {
      if (authService) return authService;

      authService = attachPlugins(
        auth({
          pachdAddress,
          channelCredentials,
          credentialMetadata,
        }),
        plugins,
      );
      return authService;
    },
    projects: () => {
      if (projectsService) return projectsService;

      projectsService = attachPlugins(
        projects({
          pachdAddress,
          channelCredentials,
          credentialMetadata,
        }),
        plugins,
      );
      return projectsService;
    },
    attachCredentials: ({
      authToken = '',
      projectId = '',
    }: Pick<ClientArgs, 'authToken' | 'projectId'>) => {
      credentialMetadata.set('authn-token', authToken);
      credentialMetadata.set('project-id', projectId);
    },
    admin: () => {
      if (adminService) return adminService;

      adminService = attachPlugins(
        admin({
          pachdAddress,
          channelCredentials,
          credentialMetadata,
        }),
        plugins,
      );
      return adminService;
    },
  };

  return methods;
};

export default client;
