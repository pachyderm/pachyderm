import '@testing-library/cypress/add-commands';

// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })
Cypress.Commands.add('resetApp', () => {
  cy.visit('?reset');
});

Cypress.Commands.add('isAppReady', () => {
  cy.get('[id="main"]').should('exist');
  // Wait for the splash screen to disappear
  cy.get('[id="jupyterlab-splash"]').should('not.exist');
  // Wait for main content to load
  cy.get('li.lm-TabBar-tab.p-TabBar-tab.lm-mod-current').should('exist');
});

Cypress.Commands.add('jupyterlabCommand', (command) => {
  cy.window().then((window) => {
    const app = window.jupyterlab ?? window.jupyterapp;
    app.commands.execute(command);
  });
});

Cypress.Commands.add('openMountPlugin', () => {
  cy.get(buildTabSelector('pachyderm-mount')).then((tab) => {
    if (!tab.hasClass('lm-mod-current')) {
      cy.get(buildTabSelector('pachyderm-mount')).click();
    }
  });
});

Cypress.Commands.add('unmountAllRepos', (command) => {
  cy.request('PUT', 'http://localhost:8888/pachyderm/v2/_unmount_all');
});

//Taken from galata
function buildTabSelector(id) {
  return `.lm-TabBar.jp-SideBar .lm-TabBar-content li.lm-TabBar-tab[data-id="${id}"]`;
}
