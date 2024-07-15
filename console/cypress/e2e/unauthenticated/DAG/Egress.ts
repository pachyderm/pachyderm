describe('Egress', () => {
  before(() => {
    cy.deleteProjectReposAndPipelines('egress');
    cy.multiLineExec(`
    pachctl create project egress
    pachctl create repo data --project egress
    echo '{"pipeline": {"name": "db"},  "egress": {"sqlDatabase": {"url": "db://url","secret": {"name": "name", "key": "key"}}} ,"transform": {"cmd": ["sh"],"stdin": [""]},"input": {"pfs": {"glob": "/*","repo": "data"}}}' | pachctl create pipeline --project egress
    echo '{"pipeline": {"name": "storage"},  "egress": {"objectStorage": {"url": "os://url"}} ,"transform": {"cmd": ["sh"],"stdin": [""]},"input": {"pfs": {"glob": "/*","repo": "data"}}}' | pachctl create pipeline --project egress
    echo '{"pipeline": {"name": "s3"},"s3Out": true,  "egress": {"objectStorage": {"url": "os://url"}} ,"transform": {"cmd": ["sh"],"stdin": [""]},"input": {"pfs": {"glob": "/*","repo": "data"}}}' | pachctl create pipeline --project egress
  `);
  });

  beforeEach(() => {
    cy.visit('/lineage/egress')
  });

  after(() => {
    cy.deleteProjectReposAndPipelines('egress');
  });

  it('should show egress node to DB', () => {
    cy.get('g[id="GROUP_1a332885c5fd1bb98311171250b9ce9c01be8a15"]')
      .should('exist')
      .findByText('Egress');

    cy.findByRole('button', {
      name: 'GROUP_1a332885c5fd1bb98311171250b9ce9c01be8a15 egress url',
      timeout: 10000,
    })
      .should('exist')
      .findByText('db://url');

    cy.get('g[id="GROUP_1a332885c5fd1bb98311171250b9ce9c01be8a15_to"]')
      .should('exist')
      .findByText('Egresses to DB');
  });

  it('should show egress node to storage', () => {
    cy.get('g[id="GROUP_3a0c236177a720b154fd3f0f303d314c5d6c6ae5"]')
      .should('exist')
      .findByText('Egress');

    cy.findByRole('button', {
      name: 'GROUP_3a0c236177a720b154fd3f0f303d314c5d6c6ae5 egress url',
      timeout: 10000,
    })
      .should('exist')
      .findByText('os://url');

    cy.get('g[id="GROUP_3a0c236177a720b154fd3f0f303d314c5d6c6ae5_to"]')
      .should('exist')
      .findByText('Egresses to storage');
  });

  it('should show egress node to S3', () => {
    cy.get('g[id="GROUP_f8059415e742a1d56d3b92e4831caed005fe3ce7"]')
      .should('exist')
      .findByText('Egress');

    cy.findByRole('button', {
      name: 'GROUP_f8059415e742a1d56d3b92e4831caed005fe3ce7 egress url',
      timeout: 10000,
    })
      .should('exist')
      .findByText('s3://s3');

    cy.get('g[id="GROUP_f8059415e742a1d56d3b92e4831caed005fe3ce7_to"]')
      .should('exist')
      .findByText('Egresses to S3');
  });

  it('should show egress node with a global id filter', () => {
    cy.get('g[id="GROUP_f8059415e742a1d56d3b92e4831caed005fe3ce7"]')
      .should('exist');

    cy.findByRole('button', {name: /previous jobs/i}).click();
    cy.findAllByTestId('SearchResultItem__container').eq(0).click();

    cy.findByRole('button', {
      name: 'GROUP_f8059415e742a1d56d3b92e4831caed005fe3ce7 egress url',
      timeout: 10000,
    })
      .should('exist')
      .findByText('s3://s3');

    cy.get('g[id="GROUP_f8059415e742a1d56d3b92e4831caed005fe3ce7_to"]')
      .should('exist')
      .findByText('Egresses to S3');
  });
});
