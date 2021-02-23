/* eslint-disable @typescript-eslint/naming-convention */
import {ApolloError} from 'apollo-server-express';
import {Issuer} from 'openid-client';

import {MutationResolvers} from 'generated/types';
import client from 'grpc/client';

const {
  ISSUER_URI: issuerUri,
  OAUTH_CLIENT_ID: clientId,
  OAUTH_CLIENT_SECRET: clientSecret,
  OAUTH_REDIRECT_URI: redirectUri,
} = process.env;

interface AuthResolver {
  Mutation: {
    exchangeCode: MutationResolvers['exchangeCode'];
  };
}

const authResolver: AuthResolver = {
  Mutation: {
    exchangeCode: async (_field, {code}, {pachdAddress = ''}) => {
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

        return client(pachdAddress).auth().authenticate(id_token);
      } catch (e) {
        throw new ApolloError(e);
      }
    },
  },
};

export default authResolver;
