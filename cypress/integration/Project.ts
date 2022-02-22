describe('Project', () => {
  before(() => {
    cy.setupProject().visit('/');
  })
  beforeEach(() => {
    cy.findAllByText('View Project').eq(0).click();
  });
  afterEach(() => {
    cy.visit('/')
  })

  after(() => {
    cy.deleteReposAndPipelines().logout();
  })
  it('should navigate to the project page', () => {
    cy.findByText('Jobs', {timeout: 8000});
    cy.findByText('Reset Canvas');
    cy.findByText('images');
    cy.findAllByText('edges').should('have.length', 2);
  });

  it('should show the correct number of commits', () => {

    cy.findByText('images').click();
    cy.findAllByText('Committed', {exact: false}).should('have.length', 2);
    cy.findAllByText('(57.27 KB)', {exact: false}).should('have.length', 2);
    cy.findByRole('button', {name: 'Close'}).click();
    cy.findAllByText('edges').eq(0).click();
    cy.findAllByText('Committed', {exact: false}).should('have.length', 1);
    cy.findByText('(22.22 KB)', {exact: false, timeout: 60000});

  });

  it('should enable the pipeline and repo deletion buttons when all downstream pipelines and repos are deleted', () => {
    cy.exec('jq -r .pachyderm version.json').then(res => {
      cy.exec(`pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/v${res.stdout}/examples/opencv/montage.json`);
    });

    cy.findByTestId('DAGView__centerSelections', {timeout: 10000}).click();


    cy.findByText('images').click();
    cy.findByTestId('DeleteRepoButton__link').should('be.disabled');
    cy.findByLabelText('Close').click();

    cy.findAllByText('edges').eq(1).click()  
    cy.findByTestId('DeletePipelineButton__link').should('be.disabled');
    cy.findByLabelText('Close').click();

    cy.findAllByText('montage').eq(1).click();
    cy.findByTestId('DeletePipelineButton__link').click();
    cy.findByTestId('ModalFooter__confirm').click();
    cy.get('[data-test-id="ModalFooter__confirm"').should('not.exist');
    cy.waitUntil(() => cy.findAllByText('montage').its.length === 0);
    cy.url().should('not.contain', '/pipelines');

    cy.findByText('images').click({force: true});
    cy.findByTestId('DeleteRepoButton__link').should('be.disabled');
    cy.findByLabelText('Close').click();

    
    cy.findAllByText('edges').eq(1).click();
    cy.findByTestId('DeletePipelineButton__link').click();
    cy.findByTestId('ModalFooter__confirm').click();
    cy.get('[data-test-id="ModalFooter__confirm"').should('not.exist');
    cy.waitUntil(() => cy.findAllByText('edges').its.length === 0);
    cy.url().should('not.contain', '/pipelines');

    cy.findByText('images').click({force: true});
    cy.findByTestId('DeleteRepoButton__link').should('not.be.disabled');
  })
});
