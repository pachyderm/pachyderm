describe('help', () => {
  beforeEach(() => {
    cy.visit('/lab');
  });

  // afterEach(() => {
  //   cy.logout();
  // });

  it('Should open the pachyderm docs from the help menu.', () => {
    cy.findAllByText('Help').first().click();
    cy.findByText('Pachyderm Docs');
  });
});
