import {Request, Response} from 'express';
import {verify} from 'jsonwebtoken';
import {Issuer} from 'openid-client';

export const getAuthAccount = async (req: Request, res: Response) => {
  const idToken = req.body.idToken;

  if (!idToken || typeof idToken !== 'string') {
    return res
      .json({
        message: 'Bad Request',
        details: ['An "idToken" is required.'],
      })
      .status(400);
  }

  try {
    const issuer = await getTokenIssuer();
    const keystore = await issuer.keystore();

    verify(
      idToken,
      (jwtHeader, callback) => {
        const key = keystore.get({kid: jwtHeader.kid});

        if (!key) {
          callback(new Error('Could not find matching key'));
        } else {
          callback(null, key.toPEM());
        }
      },
      (error, decoded) => {
        if (error) {
          return res
            .json({
              message: 'Authentication Error',
              details: [String(error)],
            })
            .status(401);
        }

        // The decoded value is a generic 'object' type
        // not sure there is a good way to garauntee typing in this case,
        // as the scope is set by the client
        const tokenContents = decoded as {
          sub: string;
          email: string;
          name?: string;
        };

        return res.json({
          // NOTE: We may need to return email for both of these fields depending
          // on what we use the ID for. "sub" is technically correct, as it's assigned
          // by the downstream IDP. However, I believe pachyderm references users by email.
          id: tokenContents.sub,
          email: tokenContents.email,
          name: tokenContents.name,
        });
      },
    );
  } catch (error) {
    return res
      .json({
        message: 'Authentication Error',
        details: [String(error)],
      })
      .status(401);
  }
};

export const exchangeCode = async (req: Request, res: Response) => {
  const {OAUTH_REDIRECT_URI: redirectUri = ''} = process.env;
  const code = req.body.code;

  if (!code || typeof code !== 'string') {
    return res
      .json({
        message: 'Bad Request',
        details: ['A "code" is required.'],
      })
      .status(400);
  }

  try {
    const oidcClient = await getOIDCClient();
    let idToken = '';

    try {
      idToken =
        (
          await oidcClient.callback(redirectUri, {
            code,
          })
        ).id_token || '';
    } catch (error) {
      return res
        .json({
          message: 'Authentication Error',
          details: [String(error)],
        })
        .status(401);
    }

    return res.json({idToken});
  } catch (error) {
    return res
      .json({
        message: 'Authentication Error',
        details: [String(error)],
      })
      .status(401);
  }
};

export const getAuthConfig = async (req: Request, res: Response) => {
  let issuer;

  try {
    issuer = await getTokenIssuer();
  } catch (error) {
    return res
      .json({
        message: 'Authentication Error',
        details: ['Unable to connect to authorization issuer.', String(error)],
      })
      .status(401);
  }

  try {
    const authUrl = new URL(issuer.metadata.authorization_endpoint || '');
    const config = {
      authEndpoint: authUrl.pathname,
      clientId: process.env.OAUTH_CLIENT_ID || '',
      pachdClientId: process.env.OAUTH_PACHD_CLIENT_ID || '',
    };

    if (!authUrl.pathname || !config.clientId || !config.pachdClientId) {
      return res
        .json({
          message: 'Authentication Error',
          details: ['Issuer is misconfigured.'],
        })
        .status(401);
    }

    return res.json(config);
  } catch (error) {
    return res
      .json({
        message: 'Authentication Error',
        details: [
          'Invalid Auth Config. Your IDP may be misconfigured.',
          String(error),
        ],
      })
      .status(401);
  }
};

const WELL_KNOWN_URL = '.well-known/openid-configuration';

const getTokenIssuer = () => {
  const {ISSUER_URI: issuerUri = ''} = process.env;

  // we need to add well-known to the uri to prevent the openid-client library from trying
  // additional uris on failed requests
  const issuerUrl = new URL(issuerUri);

  if (issuerUrl.pathname.endsWith('/')) {
    issuerUrl.pathname += WELL_KNOWN_URL;
  } else {
    issuerUrl.pathname += `/${WELL_KNOWN_URL}`;
  }

  return Issuer.discover(issuerUrl.toString());
};

const getOIDCClient = async () => {
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
