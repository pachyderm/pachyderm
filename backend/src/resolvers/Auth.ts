/* eslint-disable @typescript-eslint/naming-convention */
import {ApolloError, AuthenticationError} from 'apollo-server-express';

import {getOIDCClient, getTokenIssuer} from '@dash-backend/lib/auth';
import {UnauthenticatedContext} from '@dash-backend/lib/types';
import {AuthConfig, MutationResolvers, QueryResolvers} from '@graphqlTypes';

interface AuthResolver {
  Query: {
    account: QueryResolvers['account'];
    authConfig: QueryResolvers<UnauthenticatedContext>['authConfig'];
  };
  Mutation: {
    exchangeCode: MutationResolvers<UnauthenticatedContext>['exchangeCode'];
  };
}

const authResolver: AuthResolver = {
  Query: {
    account: async (_field, _args, {account}) => {
      return account;
    },
    authConfig: async (_field, _args, {log}) => {
      try {
        const issuer = await getTokenIssuer();

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
          throw new ApolloError('Issuer is misconfigured');
        }

        return config;
      } catch (e) {
        throw new ApolloError(String(e));
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

        const {id_token: idToken = ''} = await oidcClient.callback(
          redirectUri,
          {
            code,
          },
        );

        const pachToken = await pachClient.auth().authenticate(idToken);

        return {pachToken, idToken};
      } catch (e) {
        throw new AuthenticationError(String(e));
      }
    },
  },
};

export default authResolver;
