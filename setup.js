const {exec} = require('child_process');
const {writeFile, readFileSync} = require('fs');
const {exit} = require('process');
const {createInterface} = require('readline');
const {parse, stringify} = require('envfile');

const LOCAL_PACHD_ADDRESS = 'localhost:30650';

// -------- utilities ----------
const executePachCommand = (command) => {
  return new Promise((res, rej) => {
    exec(`pachctl ${command}`, (err, stdout) => {
      if (err) {
        rej(err);
      } else {
        res(stdout);
      }
    });
  });
};

const writeToEnv = (variable, value) => {
  const environment = readFileSync('.env.development.local', 'utf8');
  const envObject = parse(environment) || {};

  envObject[variable] = value;
  return new Promise((res, rej) => {
    writeFile('.env.development.local', stringify(envObject), (err) => {
      if (err) {
        rej(err);
      } else {
        res();
      }
    })
  })
};

const askQuestion = (question) => {
  const readline = createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((res) => {
    readline.question(question + ' ', (answer) => {
      res(answer);
      readline.close();
    });
  });
};
// ------------------------------

const checkPachVersion = async () => {
  console.log('Checking pachd version...');
  try {
    const output = await executePachCommand('version');
    const pachdVersion = output
      .split('\n')
      .find(line => line.includes('pachd'))
      .replace('pachd', '')
      .trim();

    if (pachdVersion.startsWith('2')) {
      console.log(`${pachdVersion} ✅`);
    } else {
      console.error('You must be using a 2.0 cluster!');
      exit(1);
    }
  } catch (e) {
    console.error('Problem checking pachd version:', e);
    exit(1);
  }
}

const writePachdAddress = async () => {
  console.log('Writing pachd address to .env.development.local...');
  try {
    await writeToEnv('PACHD_ADDRESS', LOCAL_PACHD_ADDRESS);
    console.log(`${LOCAL_PACHD_ADDRESS} ✅`);
  } catch (e) {
    console.error('Problem writing pachd address to .env.development.local:', e);
    exit(1);
  }
}

const activateAuth = async () => {
  console.log('Activating auth...');
  try {
    await executePachCommand('auth activate --only-activate');
    console.log('Auth activated ✅');
  } catch (e) {
    if (String(e).includes('ID already exists')) {
      console.log('Auth was previously activated ⚠️');
    } else {
      console.error('Problem activating auth:', e);
      exit(1);
    }
  }
}

const setupClusterAdmin = async () => {
  const userEmail = process.env.PACHYDERM_AUTH_EMAIL || await askQuestion('What is the primary email associated with your connected account?');
  if(process.env.NODE_ENV === 'development') {
    writeToEnv('PACHYDERM_AUTH_EMAIL', userEmail);
  }
  if(process.env.AUTH_CLIENT !== 'github') {
    const userPassword = process.env.PACHYDERM_AUTH_PASSWORD || await askQuestion('What is the password for the associated account?');
    writeToEnv('PACHYDERM_AUTH_PASSWORD', userPassword);
  }

  console.log('Setting up cluster admin...');
  try {
    await executePachCommand(`auth set cluster clusterAdmin user:${userEmail}`);
    console.log(`${userEmail} added as a cluster admin ✅`);
  } catch (e) {
    console.error('Problem creating cluster admin:', e);
    exit(1);
  }
}

const setEnterpriseContext = async () => {
  console.log('Setting enterprise context...');

  try {
    await executePachCommand('config set active-enterprise-context `pachctl config get active-context`');
    console.log('Enterprise context set ✅');
  } catch (e) {
    console.error('Problem setting enterprise context:', e);
    exit(1);
  }
}

const configureIssuer = async () => {
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
}

const configureGithubConnector = async () => {
  const clientId = process.env.PACHYDERM_AUTH_CLIENT_ID || await askQuestion('Enter Github Client ID:');
  if (process.env.NODE_ENV === "development") {
    writeToEnv("PACHYDERM_AUTH_CLIENT_ID", clientId);
  }
  const clientSecret = process.env.PACHYDERM_AUTH_CLIENT_SECRET || await askQuestion('Enter Github Client Secret:');
  if (process.env.NODE_ENV === "development") {
    writeToEnv("PACHYDERM_AUTH_CLIENT_SECRET", clientSecret);
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
}

const configureAuth0Connector = async () => {
  const clientId = process.env.PACHYDERM_AUTH_CLIENT_ID || await askQuestion('Enter Auth0 Client ID:');
  if (process.env.NODE_ENV === "development") {
    writeToEnv("PACHYDERM_AUTH_CLIENT_ID", clientId);
  }
  const clientSecret = process.env.PACHYDERM_AUTH_CLIENT_SECRET || await askQuestion('Enter Auth0 Client Secret:');
  if (process.env.NODE_ENV === "development") {
    writeToEnv("PACHYDERM_AUTH_CLIENT_SECRET", clientSecret);
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
}

const configureOIDCProvider = async () => {
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
}

const configureDashClient = async () => {
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
}

const configurePachClient = async () => {
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
}

const setup = async () => {
  await checkPachVersion();
  await writePachdAddress();
  await activateAuth();
  await setupClusterAdmin();
  await configureIssuer();
  if (process.env.AUTH_CLIENT == 'github') {
    await configureGithubConnector();
  } else {
    await configureAuth0Connector();
  }
  await configureOIDCProvider();
  await configureDashClient();
  await configurePachClient();
}

setup();
