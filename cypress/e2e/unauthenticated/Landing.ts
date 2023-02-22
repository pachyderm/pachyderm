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
    cy.deleteReposAndPipelines();
  })

  it('should show default project info', () => {
    cy.findByRole('heading', {
      name: /default/i,
    }).click();
    cy.findByRole('heading', {
      name: 'Project Preview',
    });

    cy.findByText('Total No. of Repos/Pipelines');
    cy.findByText('2/1');

    cy.findByText('Total Data Size');

    cy.findByText('Pipeline Status');
    cy.findByText('Last Job');
  });
});
