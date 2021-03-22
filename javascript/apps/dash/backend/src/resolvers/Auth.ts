/* eslint-disable @typescript-eslint/naming-convention */
import {AuthenticationError} from 'apollo-server-express';

import client from '@dash-backend/grpc/client';
import {getOIDCClient} from '@dash-backend/lib/auth';
import log from '@dash-backend/lib/log';
import {UnauthenticatedContext} from '@dash-backend/lib/types';
import authenticated from '@dash-backend/middleware/authenticated';
import {MutationResolvers, QueryResolvers} from '@graphqlTypes';

interface AuthResolver {
  Query: {
    account: QueryResolvers['account'];
  };
  Mutation: {
    exchangeCode: MutationResolvers<UnauthenticatedContext>['exchangeCode'];
  };
}

const authResolver: AuthResolver = {
  Query: {
    account: authenticated(async (_field, _args, {account}) => {
      return account;
    }),
  },
  Mutation: {
    exchangeCode: async (_field, {code}, {pachdAddress = ''}) => {
      const {OAUTH_REDIRECT_URI: redirectUri = ''} = process.env;

      try {
        const oidcClient = await getOIDCClient();

        log.info({
          eventSource: 'exchangeCode resolver',
          event: 'retreiving token from issuer',
        });

        const {id_token: idToken = ''} = await oidcClient.callback(
          redirectUri,
          {
            code,
          },
        );

        const pachToken = await client({pachdAddress, log})
          .auth()
          .authenticate(idToken);

        return {pachToken, idToken};
      } catch (e) {
        throw new AuthenticationError(e);
      }
    },
  },
};

export default authResolver;
