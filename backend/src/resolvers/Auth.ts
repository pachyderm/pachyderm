import {AuthenticationError} from 'apollo-server-express';

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
  },
};

export default authResolver;
