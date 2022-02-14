describe('Landing', () => {
  beforeEach(() => {
    cy.login();
  });

  afterEach(() => {
    cy.logout();
  });

  it('should show deafult project info', () => {
    cy.findByText('Default Pachyderm project.').click();
    cy.findByText('Project Preview');
    cy.findByText('Total No. of Repos/Pipelines');
    cy.findByText('2/1');
    cy.findByText('Pipeline Status');
  });
});
