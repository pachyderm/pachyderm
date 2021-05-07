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
// Cypress.Commands.add("login", (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add("drag", { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add("dismiss", { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite("visit", (originalFn, url, options) => { ... })

const API_KEY = Cypress.env('API_KEY') || '';
let dashUrl = '';

Cypress.Commands.add('login', () => {
  localStorage.setItem('apiKey', API_KEY);

  return cy.visit('/', {
    onBeforeLoad: (win) => {
      cy.stub(win, 'open', (args) => {
        dashUrl = args;
      });
    }
  });
});

Cypress.Commands.add('openDash', (workspaceName) => {
  cy.findByLabelText(`Workspace ${workspaceName} actions`).click();
  cy.findAllByText('Dash').filter(':visible').click();
  cy.window().its('open').should('be.called').then(() => {
    cy.writeFile('/tmp/dashUrl.txt', dashUrl);
  });
});

Cypress.Commands.add('visitDash', () => {
  cy.readFile('/tmp/dashUrl.txt').then((url) => {
    // This guarantees that cypress will retry visiting the url 4 times
    cy.visit(url, { retryOnStatusCodeFailure: true });
  
    cy.get('input', { timeout: 10000 }).should('have.attr', 'placeholder', 'Search Pachyderm');
  });
});

Cypress.Commands.add('createWorkspace', (workspaceName) => {
  cy.findByTestId('TableViewHeader__button').click();
  cy.findByTestId('WorkspaceCreateModal__name').clear();
  cy.findByTestId('WorkspaceCreateModal__name').type(workspaceName);
  cy.findByTestId('WorkspaceCreateModal__dataConfirmation').click();
  cy.findByTestId('WorkspaceCreateModal__confirm').click()
  cy.findByText(workspaceName);

  return cy.findAllByTestId('WorkspaceTableRow__row')
      .filter(`:contains(${workspaceName})`)
      .contains('Ready', {timeout: 1200000});
});

Cypress.Commands.add('deleteWorkspace', (workspaceName) => {
  cy.findAllByTestId('WorkspaceTableRow__row').filter(`:contains(${workspaceName})`).within(() => {
    cy.findByTestId('WorkspaceTableRow__dropdownButton').click();
    cy.findAllByTestId('WorkspaceTableRow__delete').filter(':visible').click();
  });
  cy.findByTestId('WorkspaceDeleteModal__confirm').click();
  
  return cy.findAllByTestId('WorkspaceTableRow__row')
    .filter(`:contains(${workspaceName})`)
    .contains(/Deleting/)
    .should('exist');
});

Cypress.Commands.add('changeOrg', (orgName) => {
  cy.findByTestId('Orgs__dropdownButton').click();
  cy.findAllByTestId('DropdownMenuItem__button')
    .contains(orgName)
    .click();
});

Cypress.Commands.add('logout', () => {
  cy.findByLabelText('Avatar').click();
  
  return cy.findByText('Logout').click();
});
