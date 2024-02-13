describe('examples', () => {
  beforeEach(() => {
    cy.resetApp();
    cy.isAppReady();
  });

  it('Should open the pachyderm tutorial example notebook.', () => {
    cy.jupyterlabCommand('filebrowser:create-main-launcher');
    cy.findByText('Pachyderm Examples');
    cy.findAllByText('Intro to Pachyderm').first().click();
    cy.contains('h1', 'Intro to Pachyderm Tutorial Header');
  });

  it('Should open the pachyderm mount example notebook.', () => {
    cy.jupyterlabCommand('filebrowser:create-main-launcher');
    cy.findByText('Pachyderm Examples');
    cy.findAllByText('Mounting Data Repos').first().click();
    cy.contains('h1', 'Mounting Data Repos in Notebooks Header');
  });
});
