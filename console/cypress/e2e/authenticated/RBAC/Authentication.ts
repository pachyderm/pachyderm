before(() => {
  cy.logout();
});

after(() => {
  cy.visit('/');
});

it('should route users back to their original page after logging in', () => {
  cy.visit('/lineage/default', {timeout: 12000});

  cy.findByRole('textbox', {timeout: 12000}).type('admin');
  cy.findByLabelText('Password').type('password');
  cy.findByRole('button', {name: /login/i, timeout: 12000}).click();

  cy.window().then((win) => {
    cy.waitUntil(() => Boolean(win.localStorage.getItem('auth-token')), {
      errorMsg: 'Auth-token was not set in localstorage',
    });
  });

  cy.url().should('include', '/lineage/default');
});
