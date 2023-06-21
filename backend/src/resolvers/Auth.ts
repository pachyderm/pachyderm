import {AuthenticationError} from 'apollo-server-express';

import {getOIDCClient, getTokenIssuer} from '@dash-backend/lib/auth';
import {
  toGQLPermissionType,
  toProtoResourceType,
} from '@dash-backend/lib/gqlEnumMappers';
import {UnauthenticatedContext} from '@dash-backend/lib/types';
import {AuthConfig, MutationResolvers, QueryResolvers} from '@graphqlTypes';
interface AuthResolver {
  Query: Required<
    Pick<
      QueryResolvers,
      'account' | 'getPermissions' | 'getRoles' | 'getAuthorize'
    > &
      Pick<QueryResolvers<UnauthenticatedContext>, 'authConfig'>
  >;
  Mutation: Pick<
    MutationResolvers<UnauthenticatedContext>,
    'exchangeCode' | 'modifyRoles'
  >;
}

const authResolver: AuthResolver = {
  Query: {
    account: async (_field, _args, {account}) => {
      return account;
    },
    authConfig: async (_field, _args, {log, account}) => {
      if (account && account.id === 'unauthenticated') {
        return {authEndpoint: '', clientId: '', pachdClientId: ''};
      }

      let issuer;
      try {
        issuer = await getTokenIssuer();
      } catch (e) {
        log.error({eventSource: 'auth issuer'}, String(e));
        throw new AuthenticationError(
          'Unable to connect to authorization issuer',
        );
      }

      try {
        const authUrl = new URL(issuer.metadata.authorization_endpoint || '');

        const config: AuthConfig = {
          authEndpoint: authUrl.pathname,
          clientId: process.env.OAUTH_CLIENT_ID || '',
          pachdClientId: process.env.OAUTH_PACHD_CLIENT_ID || '',
        };

        if (!authUrl.pathname || !config.clientId || !config.pachdClientId) {
          log.error(
            {eventSource: 'authUrl resolver', meta: {config}},
            'Issuer is missing authorization_endpoint configuration',
          );
          throw new AuthenticationError('Issuer is misconfigured');
        }

        return config;
      } catch (e) {
        log.error(
          {eventSource: 'authConfig'},
          'Issuer is missing authorization_endpoint configuration',
        );
        throw new AuthenticationError(
          'Invalid Auth Config. Your IDP may be misconfigured.',
          {sourceError: e},
        );
      }
    },
    getPermissions: async (_field, {args: {resource}}, {pachClient}) => {
      return await pachClient.auth().getPermissions({resource});
    },
    getRoles: async (_field, {args: {resource}}, {pachClient}) => {
      const roleBindings = await pachClient.auth().getRoleBinding({
        resource: {
          name: resource.name,
          type: toProtoResourceType(resource.type),
        },
      });
      return {
        roleBindings: roleBindings.binding?.entriesMap.map(
          ([name, rolesMap]) => {
            return {
              principal: name,
              roles: rolesMap.rolesMap.map(([role]) => role),
            };
          },
        ),
      };
    },
    getAuthorize: async (
      _field,
      {args: {permissionsList, resource}},
      {pachClient},
    ) => {
      try {
        const res = await pachClient.auth().authorize({
          permissionsList,
          resource,
        });

        const satisfiedList = res.satisfiedList.map((el) =>
          toGQLPermissionType(el),
        );
        const missingList = res.missingList.map((el) =>
          toGQLPermissionType(el),
        );
        const gqlSafeRes = {...res, satisfiedList, missingList};
        return gqlSafeRes;
      } catch {
        return {
          satisfiedList: [],
          missingList: [],
          principal: '',
          authorized: null,
        };
      }
    },
  },
  Mutation: {
    exchangeCode: async (_field, {code}, {pachClient, log}) => {
      const {OAUTH_REDIRECT_URI: redirectUri = ''} = process.env;

      try {
        const oidcClient = await getOIDCClient();

        log.info(
          {
            eventSource: 'exchangeCode resolver',
          },
          'retreiving token from issuer',
        );

        let idToken = '';
        try {
          idToken =
            (
              await oidcClient.callback(redirectUri, {
                code,
              })
            ).id_token || '';
        } catch (e) {
          log.error(
            {
              eventSource: 'oidc client',
              err: e,
            },
            String(e),
          );
          throw e;
        }

        let pachToken;
        try {
          pachToken = await pachClient.auth().authenticate(idToken);
        } catch (e) {
          log.error(
            {
              eventSource: 'pachd',
            },
            String(e),
          );
          throw e;
        }

        return {pachToken, idToken};
      } catch (e) {
        throw new AuthenticationError(String(e));
      }
    },
    modifyRoles: async (
      _field,
      {args: {resource, principal, rolesList}},
      {pachClient},
    ) => {
      await pachClient.auth().modifyRoleBinding({
        resource: {
          name: resource.name,
          type: toProtoResourceType(resource.type),
        },
        principal,
        rolesList,
      });

      return true;
    },
  },
};

export default authResolver;
