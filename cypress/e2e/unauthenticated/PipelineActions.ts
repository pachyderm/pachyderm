describe('Pipeline Actions', () => {
  before(() => {
    cy.deleteReposAndPipelines();
  });

  beforeEach(() => {
    cy.visit('/');
    cy.findAllByText(/^View(\sProject)*$/)
      .eq(0)
      .click();
  });

  afterEach(() => {
    cy.deleteReposAndPipelines();
  });

  it('should allow a user to rerun a pipeline and then stop the job', () => {
    cy.exec('pachctl create repo images')
      .exec(
        'pachctl put file images@master:liberty.png -f cypress/fixtures/liberty.png',
      )
      .exec(`pachctl create pipeline -f cypress/fixtures/edges.pipeline.json`);

    cy.visit('/lineage/default/pipelines/edges');

    cy.findByRole('heading', {name: /success/i, timeout: 30_000}).should('exist');

    cy.findByRole('button', {name: /pipeline actions/i}).click();
    cy.findByRole('menuitem', {name: /rerun pipeline/i}).click();

    cy.findByRole('radio', {name: /process all datums/}).click({force: true});
    cy.findByRole('button', {name: /rerun pipeline/i}).click();

    cy.findByRole('heading', {name: /created/i}).should('exist');

    cy.findByRole('button', {name: /stop job/i}).click();

    cy.findByRole('heading', {name: /killed/i}).should('exist');
  });

  it('should allow a user to pause and restart a pipeline', () => {
    cy.exec('pachctl create repo images')
      .exec(
        'pachctl put file images@master:liberty.png -f cypress/fixtures/liberty.png',
      )
      .exec(`pachctl create pipeline -f cypress/fixtures/edges.pipeline.json`);

    cy.visit('/lineage/default/pipelines/edges/info');

    cy.findByRole('heading', {name: /running/i, timeout: 30_000}).should('exist');

    cy.findByRole('button', {name: /pipeline actions/i}).click();
    cy.findByRole('menuitem', {name: /pause pipeline/i}).click();

    cy.findByRole('heading', {name: /paused/i, timeout: 20_000}).should('exist');

    cy.findByRole('button', {name: /pipeline actions/i}).click();
    cy.findByRole('menuitem', {name: /restart pipeline/i}).click();

    cy.findByRole('heading', {name: /running/i, timeout: 20_000}).should('exist');
  });

  it('should allow a user to trigger a cron pipeline', () => {
    cy.exec(`pachctl create pipeline -f cypress/fixtures/cron.pipeline.json`);

    cy.visit('/lineage/default/pipelines/cron');

    cy.findByRole('heading', {name: /success/i, timeout: 30_000}).should(
      'exist',
    );

    cy.findByRole('button', {name: /pipeline actions/i}).click();
    cy.findByRole('menuitem', {name: /run cron/i}).click();

    cy.findByRole('heading', {name: /created|running/i}).should('exist');
  });
});
