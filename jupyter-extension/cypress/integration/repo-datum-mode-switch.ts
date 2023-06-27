describe('switching between repo and datum mode', () => {
  beforeEach(() => {
    cy.resetApp();
    cy.isAppReady();
    cy.unmountAllRepos();
    cy.openMountPlugin();
    cy.findAllByText('Mount');
    cy.wait(3000);
  });

  it('should open datum mode', () => {
    cy.findByTestId('Datum__mode').click();
    cy.findAllByText('Test Datums').should('have.length', 1);
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('attr', 'placeholder')
      .should('contain', 'repo: images');
  });

  it('mounted repos appear in datum input spec and again when switching back', () => {
    cy.findAllByText('Mount').first().click();
    cy.findByTestId('ListItem__select').select('branch');
    cy.findAllByText('Mount').first().click();
    cy.findAllByText('Unmount').should('have.length', 2);

    cy.findByTestId('Datum__mode').click();
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('prop', 'value')
      .should('contain', 'cross:');
    cy.findByTestId('Datum__back').click();

    cy.findAllByText('Unmount').should('have.length', 2);
    cy.wait(3000);
    cy.findAllByText('default_images').first().click();
    cy.findAllByText('liberty.png').should('have.length', 1);
    cy.findAllByText('default_images_branch').first().click();
    cy.findAllByText('branch.png').should('have.length', 1);
  });

  it.skip('modifying input spec saves and restores it when back in datum mode', () => {
    cy.findByTestId('ListItem__select').select('branch');
    cy.findAllByText('Mount').first().click();
    cy.findAllByText('Unmount').should('have.length', 1);

    cy.findByTestId('Datum__mode').click();
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('prop', 'value')
      .should('contain', 'name: default_images_branch');
    cy.findByTestId('Datum__inputSpecInput', {timeout: 12000})
      .clear()
      .type('abcd')
      .should('contain', 'a');
    cy.findByTestId('Datum__back').click();

    cy.findAllByText('Unmount').should('have.length', 1);
    cy.findAllByText('Unmount').first().click();
    cy.findByTestId('Datum__mode').click();
    cy.findByTestId('Datum__inputSpecInput')
      .invoke('prop', 'value')
      .should('contain', 'a');
  });
});
