describe('switching between repo and datum mode', () => {
  beforeEach(() => {
    cy.resetApp();
    cy.isAppReady();
    cy.openMountPlugin();
  });

  it('should open datum mode', () => {
    cy.findAllByRole('tab').filter('.pachyderm-test-tab').click();
    cy.findAllByText('Test Datums').should('have.length', 1);
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('attr', 'placeholder')
      .should('contain', 'repo: images');
  });

  it('mounted repos appear in datum input spec and again when switching back', () => {
    cy.findByTestId('ProjectRepo-DropdownCombobox-li-default/images').click();

    cy.findAllByRole('tab').filter('.pachyderm-test-tab').click();
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('prop', 'value')
      .should('contain', 'pfs:');
    cy.findAllByRole('tab').filter('.pachyderm-explore-tab').click();

    cy.wait(3000);
    cy.get('#jupyterlab-pachyderm-browser-pfs')
      .findByText('default_images')
      .dblclick();
    cy.findAllByText('liberty.png').should('have.length', 1);
  });

  it('modifying input spec saves and restores it when back in datum mode', () => {
    cy.findByTestId('ProjectRepo-DropdownCombobox-li-default/images').click();
    cy.wait(1000);
    cy.findByTestId('Branch-DropdownCombobox-input').click();
    cy.findByTestId('Branch-DropdownCombobox-li-branch').click();

    cy.findAllByRole('tab').filter('.pachyderm-test-tab').click();
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('prop', 'value')
      .should('contain', 'name: default_images_branch');
    cy.findByTestId('Datum__inputSpecInput', {timeout: 12000})
      .clear()
      .type('abcd')
      .should('contain', 'a');
    cy.findAllByRole('tab').filter('.pachyderm-explore-tab').click();

    cy.findAllByRole('tab').filter('.pachyderm-test-tab').click();
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('prop', 'value')
      .should('contain', 'a');
  });
});
