describe('Live Automation', () => {
  const workspaceName = `test-dash-${Date.now()}`;

  beforeEach(() => {
    cy.login();
  });

  afterEach(() => {
    cy.logout();
  });

  it('should open Dash', () => {
    cy.changeOrg('hub-automation-test');
    cy.createWorkspace(workspaceName);
    cy.openDash(workspaceName);
  });

  it('should connect to Dash', () => {
    cy.visitDash();
    cy.visit('/'); // navigate back to Hub to logout and delete workspace
    cy.deleteWorkspace(workspaceName);
  });
});
