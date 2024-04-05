// ***********************************************************
// This example support/index.js is processed and
// loaded automatically before your test files.
//
// This is a great place to put global configuration and
// behavior that modifies Cypress.
//
// You can change the location of this file or turn off
// automatically serving support files with the
// 'supportFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/configuration
// ***********************************************************

// Import commands.js using ES2015 syntax:
import './commands';

// Alternatively you can use CommonJS syntax:
// require('./commands')

// This code will take a debug dump of the kuberneties enviroment when a test fails.
// The dump will be uploaded as an artifact in circle ci.
Cypress.on('fail', (error, mocha) => {
  cy.exec('bash etc/testing/circle/kube_debug.sh', {
    log: true,
  }).then((result) => {
    const now = new Date();
    const formattedDate = `${now.getFullYear()}-${(now.getMonth() + 1)
      .toString()
      .padStart(2, '0')}-${now.getDate().toString().padStart(2, '0')}`;
    const formattedTime = `${now.getHours().toString().padStart(2, '0')}-${now
      .getMinutes()
      .toString()
      .padStart(2, '0')}-${now.getSeconds().toString().padStart(2, '0')}`;

    const fileName = `output_${formattedDate}_${formattedTime}.txt`;

    cy.writeFile(
      `/tmp/cypress-debug/${fileName}`,
      `
    ${mocha.title}
    ${result.stdout}
  `,
    );
  });
  throw error;
});
