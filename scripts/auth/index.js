const {LOCAL_PACHD_ADDRESS, COMMAND_SEPARATOR} = require('./constants');
const {
  activateAuth,
  setupClusterAdmin,
  configureIssuer,
  configureGithubConnector,
  configureAuth0Connector,
  configureOIDCProvider,
  configureDashClient,
  configurePachClient,
  checkEnterprise,
  activateEnterprise,
  logRootToken,
} = require('./operations');

const {prettyCommand, executeCommands} = require('./utils');

module.exports = {
  setupAuth: async () => {
    console.log('Using', LOCAL_PACHD_ADDRESS);

    if (!(await prettyCommand(checkEnterprise)))
      await prettyCommand(activateEnterprise);

    await executeCommands([
      // writePachdAddress, // You should already have this setup.
      activateAuth,
      setupClusterAdmin,
      configureIssuer,
      process.env.AUTH_CLIENT == 'github'
        ? configureGithubConnector
        : configureAuth0Connector,
      configureOIDCProvider,
      configureDashClient,
      configurePachClient,
      logRootToken,
    ]);
  },
};
