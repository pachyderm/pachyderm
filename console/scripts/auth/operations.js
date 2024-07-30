const {exit} = require('process');
const {askQuestion, writeToEnv, executePachCommand} = require('./utils');
const {LOCAL_PACHD_ADDRESS, COMMAND_SEPARATOR} = require('./constants');

let rootToken;

module.exports = {
  writePachdAddress: async () => {
    // Depreciated
    console.log('Writing pachd address to .env.development.local...');
    try {
      await writeToEnv('PACHD_ADDRESS', LOCAL_PACHD_ADDRESS);
      console.log(`${LOCAL_PACHD_ADDRESS} ✅`);
    } catch (e) {
      console.error(
        'Problem writing pachd address to .env.development.local:',
        e,
      );
      exit(1);
    }
  },
  checkEnterprise: async () => {
    console.log('Checking enterprise state...');
    const output = await executePachCommand('license get-state');

    const active = output.match(/Pachyderm Enterprise token state: (.*)/)?.[1];
    const res = active === 'ACTIVE';
    if (res) console.log('Enterprise is active.');
    if (!res) console.log('Enterprise is not active.');
    return res;
  },
  activateEnterprise: async () => {
    const PACHYDERM_ENTERPRISE_KEY =
      process.env.PACHYDERM_ENTERPRISE_KEY ||
      (await askQuestion('What is the enterprise key?'));
    await writeToEnv('PACHYDERM_ENTERPRISE_KEY', PACHYDERM_ENTERPRISE_KEY);

    console.log('Activating enterprise...');
    try {
      await executePachCommand(
        `echo ${PACHYDERM_ENTERPRISE_KEY} | pachctl license activate`,
        false,
      );
    } catch (e) {
      console.error('Problem activating enterprise:\n', e);
      exit(1);
    }
  },
  activateAuth: async () => {
    console.log('Activating auth...');
    try {
      const output = await executePachCommand('auth activate --only-activate');
      rootToken = output.match(/Pachyderm root token:\n(.*)/)?.[1];
      console.log('Auth activated ✅');
    } catch (e) {
      if (String(e).includes('ID already exists')) {
        console.log('Auth was previously activated ⚠️');
      } else {
        console.error('Problem activating auth:', e);
        exit(1);
      }
    }
  },
  setupClusterAdmin: async () => {
    const userEmail =
      process.env.PACHYDERM_AUTH_EMAIL ||
      (await askQuestion(
        'What is the primary email associated with your connected account?',
      ));
    if (process.env.NODE_ENV === 'development') {
      writeToEnv('PACHYDERM_AUTH_EMAIL', userEmail);
    }
    if (process.env.AUTH_CLIENT !== 'github') {
      const userPassword =
        process.env.PACHYDERM_AUTH_PASSWORD ||
        (await askQuestion('What is the password for the associated account?'));
      writeToEnv('PACHYDERM_AUTH_PASSWORD', userPassword);
    }

    console.log('Setting up cluster admin...');
    try {
      await executePachCommand(
        `auth set cluster clusterAdmin user:${userEmail}`,
      );
      console.log(`${userEmail} added as a cluster admin ✅`);
    } catch (e) {
      console.error('Problem creating cluster admin:', e);
      exit(1);
    }
  },
  setEnterpriseContext: async () => {
    console.log('Setting enterprise context...');

    try {
      await executePachCommand(
        'config set active-enterprise-context `pachctl config get active-context`',
      );
      console.log('Enterprise context set ✅');
    } catch (e) {
      console.error('Problem setting enterprise context:', e);
      exit(1);
    }
  },
  configureIssuer: async () => {
    console.log('Configuring issuer...');
    try {
      await executePachCommand(`idp set-config <<EOF
      {
        "issuer": "http://localhost:30658/"
      }`);
      console.log('Issuer configured ✅');
    } catch (e) {
      console.error('Problem configuring issuer:', e);
      exit(1);
    }
  },
  configureGithubConnector: async () => {
    const clientId =
      process.env.PACHYDERM_AUTH_CLIENT_ID ||
      (await askQuestion('Enter Github Client ID:'));
    if (process.env.NODE_ENV === 'development') {
      writeToEnv('PACHYDERM_AUTH_CLIENT_ID', clientId);
    }
    const clientSecret =
      process.env.PACHYDERM_AUTH_CLIENT_SECRET ||
      (await askQuestion('Enter Github Client Secret:'));
    if (process.env.NODE_ENV === 'development') {
      writeToEnv('PACHYDERM_AUTH_CLIENT_SECRET', clientSecret);
    }

    console.log('Configuring Github connector...');
    try {
      await executePachCommand(`idp create-connector <<EOF
      {
        "id": "github",
        "name": "github",
        "type": "github",
        "config": {
          "clientID": "${clientId}",
          "clientSecret": "${clientSecret}",
          "redirectURI": "http://localhost:30658/callback",
          "org": "pachyderm"
        }
      }`);
      console.log(`Connector configured with client ID ${clientId} ✅`);
    } catch (e) {
      if (String(e).includes('ID already exists')) {
        console.log('Connector previously configured ⚠️');
      } else {
        console.log('Problem configuring connector:', e);
        exit(1);
      }
    }
  },
  configureAuth0Connector: async () => {
    const clientId =
      process.env.PACHYDERM_AUTH_CLIENT_ID ||
      (await askQuestion('Enter Auth0 Client ID:'));
    if (process.env.NODE_ENV === 'development') {
      writeToEnv('PACHYDERM_AUTH_CLIENT_ID', clientId);
    }
    const clientSecret =
      process.env.PACHYDERM_AUTH_CLIENT_SECRET ||
      (await askQuestion('Enter Auth0 Client Secret:'));
    if (process.env.NODE_ENV === 'development') {
      writeToEnv('PACHYDERM_AUTH_CLIENT_SECRET', clientSecret);
    }

    console.log('Configuring Auth0 connector...');
    try {
      await executePachCommand(`idp create-connector <<EOF
      {
        "id": "auth0",
        "name": "auth0",
        "type": "oidc",
        "config": {
          "clientID": "${clientId}",
          "clientSecret": "${clientSecret}",
          "issuer": "https://hub-e2e-testing.us.auth0.com/",
          "redirectURI": "http://localhost:30658/callback",
          "org": "pachyderm"
        }
      }`);
      console.log(`Connector configured with client ID ${clientId} ✅`);
    } catch (e) {
      if (String(e).includes('ID already exists')) {
        console.log('Connector previously configured ⚠️');
      } else {
        console.log('Problem configuring connector:', e);
        exit(1);
      }
    }
  },
  configureOIDCProvider: async () => {
    console.log('Configuring OIDC provider...');
    try {
      await executePachCommand(`auth set-config <<EOF
      {
        "issuer": "http://localhost:30658/",
        "localhost_issuer": true,
        "client_id": "pachd",
        "client_secret": "notsecret",
        "redirect_uri": "http://localhost:30657/authorization-code/callback",
        "scopes": ["email", "openid", "profile", "groups"]
      }`);
      console.log('OIDC Provider configured ✅');
    } catch (e) {
      if (String(e).includes('ID already exists')) {
        console.log('Provider previously configured ⚠️');
      } else {
        console.log('Problem configuring connector:', e);
        exit(1);
      }
    }
  },
  configureDashClient: async () => {
    console.log('Configuring Dash client...');
    try {
      const output = await executePachCommand(`idp create-client <<EOF
    {
      "id": "dash",
      "name": "dash",
      "redirect_uris": ["http://localhost:4000/oauth/callback/\?inline\=true"]
    }`);
      const secret = output.match(/"(.*?)"/)[1];
      await writeToEnv('OAUTH_CLIENT_SECRET', secret);
      console.log('Dash client configured ✅');
    } catch (e) {
      if (String(e).includes('ID already exists')) {
        console.log('Dash client was previously configured ⚠️');
      } else {
        console.log('Problem configuring dash client:', e);
        exit(1);
      }
    }
  },
  configurePachClient: async () => {
    console.log('Creating pachd client...');
    try {
      await executePachCommand(`idp create-client <<EOF
    {
      "id": "pachd",
      "name": "pachd",
      "secret": "notsecret",
      "redirect_uris": ["http://localhost:30657/authorization-code/callback"],
      "trusted_peers": ["dash"]
    }`);
      console.log('Pachd client created  ✅');
    } catch (e) {
      if (String(e).includes('ID already exists')) {
        console.log('Pachd client was previously configured ⚠️');
      } else {
        console.log('Problem configuring dash client:', e);
        exit(1);
      }
    }
  },
  logRootToken: () => {
    console.log(COMMAND_SEPARATOR);

    rootToken
      ? console.log('Pachyderm root token:', rootToken)
      : console.log(
          'No pachyderm root token was found. You likely already have authorization active.',
        );
  },
};
