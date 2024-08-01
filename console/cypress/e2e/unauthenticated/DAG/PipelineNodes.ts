describe('Pipeline Nodes', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.exec('pachctl create pipeline -f cypress/fixtures/cron.pipeline.json');
    cy.exec('pachctl create pipeline -f cypress/fixtures/spout.pipeline.json');
    cy.exec(
      'pachctl create pipeline -f cypress/fixtures/spout-service.pipeline.json',
    );
    cy.exec('pachctl create repo service-input');
    cy.exec(
      'pachctl create pipeline -f cypress/fixtures/service.pipeline.json',
    );
  });
  beforeEach(() => {
    cy.visit('/lineage/default');
  });
  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('should show CRON footer on a cron repo and the cron icon on the pipeline where the input is defined', () => {
    cy.get('g[id="GROUP_ee2b69b5f254ed544b0015caf8d79c1280fb2076"]')
      .should('exist')
      .findByText('CRON');
    cy.get(
      'g[id="GROUP_74b3c5575ca0d9fe75c78b54f965230dc814e531_cronIcon"]',
    ).should('exist');
  });

  it('should show spout footer on a spout pipeline', () => {
    cy.get('g[id="GROUP_ed4a7d2e13ffb7eb32d328e9227ea5bd7d99d249"]')
      .should('exist')
      .findByText('Spout');
  });

  it('should show spout and service footer on a spout pipeline with a service', () => {
    cy.get('g[id="GROUP_fad96f957e64fa133d3adef9b2715c9e9d7a5b65"]')
      .should('exist')
      .findByText('Spout and Service');
  });

  it('should show service footer on a service pipeline', () => {
    cy.get('g[id="GROUP_1b2e1d4dc3ed671c8db2eb91d4c626e16da2f9a0"]')
      .should('exist')
      .findByText('Service');
  });

  it('should show parallelism info in footer', () => {
    //cron pipeline
    cy.get('g[id="GROUP_74b3c5575ca0d9fe75c78b54f965230dc814e531_parallelism"]')
      .should('exist')
      .findByText('1');
  });
});
