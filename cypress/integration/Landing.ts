describe('Landing', () => {
  before(() => {
    cy.setupProject();
  })
  beforeEach(() => {
    cy.visit('/')
  });

  afterEach(() => {
    cy.visit('/');
  });

  after(() => {
    cy.deleteReposAndPipelines().logout();
  })

  it('should show deafult project info', () => {
    cy.findByText('Default Pachyderm project.').click();
    cy.findByText('Project Preview');
    cy.findByText('Total No. of Repos/Pipelines');
    cy.findByText('Total Data Size');
    cy.findByText('58.65 kB');
    cy.findByText('2/1');
    cy.findByText('Pipeline Status');
    cy.findByText('Last Job');
  });
});
