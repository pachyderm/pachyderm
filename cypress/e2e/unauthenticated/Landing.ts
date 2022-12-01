describe('Landing', () => {
  before(() => {
    cy.setupProject();
  })
  beforeEach(() => {
    cy.visit('/')
    cy.findByText('Skip tutorial').click();
  });

  afterEach(() => {
    cy.visit('/');
  });

  after(() => {
    cy.deleteReposAndPipelines();
  })

  it('should show default project info', () => {
    cy.findByText('Default').click();
    cy.findByText('Project Preview');
    cy.findByText('Total No. of Repos/Pipelines');
    cy.findByText('Total Data Size');
    cy.findByText('2/1');
    cy.findByText('Pipeline Status');
    cy.findByText('Last Job');
  });
});
