describe('Live Automation', () => {
  it('should open Dash', () => {
    const workspaceName = `test-dash-${Date.now()}`;

    // This needs to be stores in a file, as switching between
    // domains causes cypress to reload the script, thereby
    // dumping everything from memory.
    cy.writeFile('/tmp/workspaceName.txt', workspaceName);

    cy.login();
    cy.changeOrg('hub-automation-test');
    cy.createWorkspace(workspaceName);
    cy.openDash(workspaceName);
    cy.logout();
  });

  it('should connect to Dash', () => {
    cy.visitDash();
  });

  it('should cleanup workspace', () => {
    cy.login();
    cy.changeOrg('hub-automation-test');
    cy.readFile('/tmp/workspaceName.txt').then((workspaceName) => {
      cy.deleteWorkspace(workspaceName);
      cy.logout();
    });
  });
});
