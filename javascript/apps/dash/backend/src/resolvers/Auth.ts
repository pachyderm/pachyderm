/* eslint-disable @typescript-eslint/naming-convention */
import {AuthenticationError} from 'apollo-server-express';
import {Issuer} from 'openid-client';

import {MutationResolvers} from '@graphqlTypes';
import client from '@dash-backend/grpc/client';

interface AuthResolver {
  Mutation: {
    exchangeCode: MutationResolvers['exchangeCode'];
  };
}

const authResolver: AuthResolver = {
  Mutation: {
    exchangeCode: async (_field, {code}, {pachdAddress = ''}) => {
      const {
        ISSUER_URI: issuerUri,
        OAUTH_CLIENT_ID: clientId,
        OAUTH_CLIENT_SECRET: clientSecret,
        OAUTH_REDIRECT_URI: redirectUri,
      } = process.env;

      try {
        const dexIssuer = await Issuer.discover(issuerUri);

        const dexClient = new dexIssuer.Client({
          client_id: clientId,
          client_secret: clientSecret,
          redirect_uris: [redirectUri],
          response_types: ['code'],
        });

        const {id_token = ''} = await dexClient.callback(redirectUri, {
          code,
        });

        const pachToken = await client(pachdAddress)
          .auth()
          .authenticate(id_token);

        return pachToken;
      } catch (e) {
        throw new AuthenticationError(e);
      }
    },
  },
};

export default authResolver;
