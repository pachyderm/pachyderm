describe('switching between repo and datum mode', () => {
  beforeEach(() => {
    cy.resetApp();
    cy.isAppReady();
    cy.unmountAllRepos();
    cy.openMountPlugin();
    cy.findAllByText('Load');
    cy.wait(3000);
  });

  it('should open datum mode', () => {
    cy.findAllByRole('tab').filter('.pachyderm-test-tab').click();
    cy.findAllByText('Test Datums').should('have.length', 1);
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('attr', 'placeholder')
      .should('contain', 'repo: images');
  });

  it('mounted repos appear in datum input spec and again when switching back', () => {
    cy.findAllByText('Load').first().click();
    cy.findByTestId('ListItem__select').select('branch');
    cy.findAllByText('Load').first().click();
    cy.findAllByText('Unload').should('have.length', 2);

    cy.findAllByRole('tab').filter('.pachyderm-test-tab').click();
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('prop', 'value')
      .should('contain', 'cross:');
    cy.findAllByRole('tab').filter('.pachyderm-explore-tab').click();

    cy.findAllByText('Unload').should('have.length', 2);
    cy.wait(3000);
    cy.findAllByText('default_images').first().click();
    cy.findAllByText('liberty.png').should('have.length', 1);
    cy.findAllByText('default_images_branch').first().click();
    cy.findAllByText('branch.png').should('have.length', 1);
  });

  it('modifying input spec saves and restores it when back in datum mode', () => {
    cy.findByTestId('ListItem__select').select('branch');
    cy.findAllByText('Load').first().click();
    cy.findAllByText('Unload').should('have.length', 1);

    cy.findAllByRole('tab').filter('.pachyderm-test-tab').click();
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('prop', 'value')
      .should('contain', 'name: default_images_branch');
    cy.findByTestId('Datum__inputSpecInput', {timeout: 12000})
      .clear()
      .type('abcd')
      .should('contain', 'a');
    cy.findAllByRole('tab').filter('.pachyderm-explore-tab').click();

    cy.findAllByText('Unload').should('have.length', 1);
    cy.findAllByText('Unload').first().click();
    cy.findAllByRole('tab').filter('.pachyderm-test-tab').click();
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('prop', 'value')
      .should('contain', 'a');
  });
});
