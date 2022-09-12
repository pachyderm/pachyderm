describe('mount', () => {
  beforeEach(() => {
    cy.unmountAllRepos();
    cy.resetApp();
    cy.isAppReady();
    cy.openMountPlugin();
    cy.findAllByText('Mount');
  });

  it('should mount and unmount pachyderm repos', () => {
    cy.findAllByText('Mount').first().click();
    cy.findAllByText('Unmount').first().click();
    cy.findAllByText('Mount').should('have.length', 1);
  });

  it('file browser should show correct breadcrumbs', () => {
    cy.findByText('/ pfs').should('have.length', 1);
    cy.findAllByText('Mount').first().click();
    cy.findAllByText('Unmount').should('have.length', 1);
    cy.findAllByText('images').first().click();

    cy.get('[id="pachyderm-mount"] div.jp-FileBrowser-crumbs')
      .first()
      .invoke('text')
      .should('eq', '/ pfs/images/');
  });

  it("should correctly mount a repo's branch", () => {
    cy.findByTestId('ListItem__select').select('branch');
    cy.findAllByText('Mount').first().click();
    cy.findAllByText('Unmount').should('have.length', 1);
    cy.findAllByText('images').first().click();
    cy.findAllByText('branch.png').should('have.length', 1);
  });

  it('should open mounted directory in the file browser on click', () => {
    cy.findAllByText('Mount').first().click();
    cy.findAllByText('Unmount').should('have.length', 1);
    cy.findAllByText('images').first().click();
    cy.findAllByText('liberty.png').should('have.length', 1);
  });

  it('file browser should show correct right click actions', () => {
    cy.findByText('Mount').first().click();
    cy.findAllByText('Unmount').should('have.length', 1);
    cy.findAllByText('images').first().click();
    cy.findAllByText('liberty.png').first().rightclick();
    cy.get('ul.lm-Menu-content.p-Menu-content')
      .children()
      .should('have.length', 2)
      .first()
      .should('have.text', 'Open')
      .next()
      .should('have.text', 'Copy Path');
  });
});
