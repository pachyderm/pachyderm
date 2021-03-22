/* eslint-disable @typescript-eslint/naming-convention */
import {AuthenticationError} from 'apollo-server-errors';
import {GetPublicKeyOrSecret, verify} from 'jsonwebtoken';
import {Issuer} from 'openid-client';

import {Account} from '@graphqlTypes';

export const getTokenIssuer = () => {
  const {ISSUER_URI: issuerUri = ''} = process.env;

  return Issuer.discover(issuerUri);
};

export const getOIDCClient = async () => {
  const {
    OAUTH_CLIENT_ID: clientId = '',
    OAUTH_CLIENT_SECRET: clientSecret = '',
    OAUTH_REDIRECT_URI: redirectUri = '',
  } = process.env;

  const issuer = await getTokenIssuer();

  return new issuer.Client({
    client_id: clientId,
    client_secret: clientSecret,
    redirect_uris: [redirectUri],
    response_types: ['code'],
  });
};

export const getAccountFromIdToken = async (
  idToken: string,
): Promise<Account> => {
  try {
    const issuer = await getTokenIssuer();
    const keystore = await issuer.keystore();

    const getKey: GetPublicKeyOrSecret = (jwtHeader, callback) => {
      const key = keystore.get({kid: jwtHeader.kid});

      if (!key) {
        callback(new Error('Could not find matching key'));
      } else {
        callback(null, key.toPEM());
      }
    };

    return new Promise((res) => {
      verify(idToken, getKey, (err, decoded) => {
        if (err) {
          throw new AuthenticationError(err.message);
        }

        // decoded is a generic 'object' type
        // not sure there is a good way to garauntee typing in this case,
        // as the scope is set by the client
        const tokenContents = decoded as {sub: string; email: string};

        res({
          // NOTE: We may need to return email for both of these fields depending
          // on what we use the ID for. "sub" is technically correct, as it's assigned
          // by the downstream IDP. However, I believe pachyderm references users by email.
          id: tokenContents.sub,
          email: tokenContents.email,
        });
      });
    });
  } catch (e) {
    throw new AuthenticationError(e);
  }
};
