/* eslint-disable @typescript-eslint/naming-convention */
export default {
  grant_types_supported: [
    'authorization_code',
    'refresh_token',
    'urn:ietf:params:oauth:grant-type:device_code',
  ],
  response_types_supported: ['code'],
  subject_types_supported: ['public'],
  id_token_signing_alg_values_supported: ['RS256'],
  code_challenge_methods_supported: ['S256', 'plain'],
  scopes_supported: ['openid', 'email', 'groups', 'profile', 'offline_access'],
  token_endpoint_auth_methods_supported: ['client_secret_basic'],
  claims_supported: [
    'aud',
    'email',
    'email_verified',
    'exp',
    'iat',
    'iss',
    'locale',
    'name',
    'sub',
  ],
};
