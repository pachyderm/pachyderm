describe('mount', () => {
  beforeEach(() => {
    cy.resetApp();
    cy.isAppReady();
    cy.unmountAllRepos();
    cy.openMountPlugin();
    cy.findAllByText('Load');
    cy.wait(3000);
  });

  it('should mount and unmount pachyderm repos', () => {
    cy.findAllByText('Load').first().click();
    cy.findAllByText('Unload').first().click();
    cy.findAllByText('Load').should('have.length', 1);
  });

  it('file browser should show correct breadcrumbs', () => {
    cy.findAllByText('/ pfs').should('have.length', 2);
    cy.findAllByText('Load').first().click();
    cy.findAllByText('Unload').should('have.length', 1);
    cy.get('#jupyterlab-pachyderm-browser-pfs')
      .findByText('default_images')
      .dblclick();

    cy.get('[id="pachyderm-mount"] div.jp-FileBrowser-crumbs')
      .first()
      .invoke('text')
      .should('eq', '/ pfs/default_images/');
  });

  it("should correctly mount a repo's branch", () => {
    cy.findByTestId('ListItem__select').select('branch');
    cy.findAllByText('Load').first().click();
    cy.findAllByText('Unload').should('have.length', 1);
    cy.get('#jupyterlab-pachyderm-browser-pfs')
      .findByText('default_images_branch')
      .dblclick();
    cy.findAllByText('branch.png').should('have.length', 1);
  });

  it('should open mounted directory in the file browser on click', () => {
    cy.findAllByText('Load').first().click();
    cy.findAllByText('Unload').should('have.length', 1);
    cy.get('#jupyterlab-pachyderm-browser-pfs')
      .findByText('default_images')
      .dblclick();
    cy.findAllByText('liberty.png').should('have.length', 1);
  });

  it('file browser should show correct right click actions', () => {
    cy.findAllByText('Load').first().click();
    cy.findAllByText('Unload').should('have.length', 1);
    cy.get('#jupyterlab-pachyderm-browser-pfs')
      .findByText('default_images')
      .dblclick();
    cy.findAllByText('liberty.png').first().rightclick();
    cy.get('ul.lm-Menu-content.p-Menu-content')
      .children()
      .should('have.length', 3)
      .first()
      .should('have.text', 'Open')
      .next()
      .should('have.text', 'Copy Path')
      .next()
      .should('have.text', 'Download');
  });

  it('file browser should have loading attribute', () => {
    cy.findAllByText('Load').first().click();
    cy.findAllByText('Unload').should('have.length', 1);
    cy.get('#jupyterlab-pachyderm-browser-pfs')
      .findByText('default_images')
      .dblclick();
    cy.get('ul.jp-DirListing-content[loading]').should('have.length', 2);
  });
});
