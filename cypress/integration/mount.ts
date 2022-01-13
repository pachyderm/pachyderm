describe('mount', () => {
  beforeEach(() => {
    cy.visit('/lab');
    cy.isAppReady();
  });

  it('should mount and unmount pachyderm repos', () => {
    cy.findAllByTitle('Pachyderm Mount').first().click();
    cy.findAllByText('Mount').first().click();
    cy.findAllByText('Unmount').first().click();
    cy.findAllByText('Mount').should('have.length', 1);
  });
});
