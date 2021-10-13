describe('Landing', () => {
  beforeEach(() => {
    cy.visit('http://localhost:4000');
  });


  it('should show top-level project info', () => {
    cy.findByText('Solar Power Data Logger Team Collab').click();
    cy.findByText('Project Preview');
    cy.findByText('Total Data Size');
    cy.findByText('607.28 KB');
    cy.findByText('Total No. of Repos/Pipelines');
    cy.findByText('2/1');
    cy.findByText('Pipeline Status');
    cy.findByText('Last 2 Jobs');
  });

  it('should navigate to the project page', () => {
    cy.findAllByText('View Project').eq(1).click();
    cy.findByText('Show Jobs');
    cy.findByText('Reset Canvas');
    cy.findByText('cron');
  });
});
