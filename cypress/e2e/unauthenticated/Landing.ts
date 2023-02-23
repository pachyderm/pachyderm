describe('Landing', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.setupProject();
    cy.exec('pachctl delete project new-project', {failOnNonZeroExit: false});
  });
  beforeEach(() => {
    cy.visit('/');
  });

  afterEach(() => {
    cy.visit('/');
  });
  after(() => {
    cy.deleteReposAndPipelines();
    cy.exec('pachctl delete project new-project', {failOnNonZeroExit: false});
  });

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

  it('should create a new project', () => {
    cy.findByRole('button', {
      name: /create project/i,
      timeout: 12000,
    }).click();

    cy.findByRole('textbox', {
      name: /name/i,
      exact: false,
    }).type('new-project');

    cy.findByRole('textbox', {
      name: /description/i,
      exact: false,
    }).type('New project description');

    cy.findByRole('button', {
      name: /create/i,
    }).click();

    cy.findByRole('heading', {
      name: /new-project/i,
      timeout: 15000,
    });

    cy.findByText('New project description');
  });
});
