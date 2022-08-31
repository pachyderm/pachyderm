describe('Project', () => {
  before(() => {
    cy.setupProject().visit('/');
  })

  beforeEach(() => {
    cy.findByText('Skip tutorial').click();
    cy.findAllByText(/^View(\sProject)*$/).eq(0).click();
  });

  afterEach(() => {
    cy.visit('/')
  })

  after(() => {
    cy.deleteReposAndPipelines();
  })

  it('should enable the pipeline and repo deletion buttons when all downstream pipelines and repos are deleted', () => {
    cy.exec('jq -r .pachyderm version.json').then(res => {
      cy.exec(`pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/v${res.stdout}/examples/opencv/montage.json`);
    });

    cy.findByTestId('DAGView__centerSelections', {timeout: 10000}).click();

    cy.findByText('images').click();
    cy.findByTestId('DeleteRepoButton__link').should('be.disabled');
    cy.findByLabelText('Close').click();

    cy.get("#GROUP_edges").within(() => cy.findByText('Pipeline').click());
    cy.findByTestId('DeletePipelineButton__link').should('be.disabled');
    cy.findByLabelText('Close').click();

    cy.get("#GROUP_montage").within(() => cy.findByText('Pipeline').click());
    cy.findByTestId('DeletePipelineButton__link').click();
    cy.findByTestId('ModalFooter__confirm').click();
    cy.get('[data-test-id="ModalFooter__confirm"').should('not.exist');
    cy.waitUntil(() => cy.findAllByText('montage').its.length === 0);
    cy.url().should('not.contain', '/pipelines');

    cy.findByText('images').click({force: true});
    cy.findByTestId('DeleteRepoButton__link').should('be.disabled');
    cy.findByLabelText('Close').click();

    
    cy.get("#GROUP_edges").within(() => cy.findByText('Pipeline').click());
    cy.findByTestId('DeletePipelineButton__link').click();
    cy.findByTestId('ModalFooter__confirm').click();
    cy.get('[data-test-id="ModalFooter__confirm"').should('not.exist');
    cy.waitUntil(() => cy.findAllByText('edges').its.length === 0);
    cy.url().should('not.contain', '/pipelines');

    cy.findByText('images').click({force: true});
    cy.findByTestId('DeleteRepoButton__link').should('not.be.disabled');
  })
});
