describe('Project', () => {
  before(() => {
    cy.setupProject().visit('/');
  });

  beforeEach(() => {
    cy.findAllByText(/^View(\sProject)*$/)
      .eq(0)
      .click();
  });

  afterEach(() => {
    cy.visit('/');
  });

  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('should enable the pipeline and repo deletion buttons when all downstream pipelines and repos are deleted', () => {
    cy.exec('jq -r .pachReleaseCommit version.json').then((res) => {
      cy.exec(
        `pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/${res.stdout}/examples/opencv/montage.pipeline.json`,
      );
    });

    // wait for jobs to finish to reduce pachd strain
    cy.findByText('Jobs').click();
    cy.findByLabelText('expand filters').click();
    cy.findAllByText('Success').filter(':visible').last().click();
    cy.findAllByTestId('RunsList__row', {timeout: 60000}).should(
      'have.length',
      2,
    );
    cy.findByText('DAG').click();

    cy.findByText('images').click();
    cy.findByTestId('DeleteRepoButton__link').should('be.disabled');
    cy.findByRole('button', {name: 'Close sidebar'}).click();

    cy.get('#GROUP_edges').within(() => cy.findByText('Pipeline').click());
    cy.findByTestId('DeletePipelineButton__link').should('be.disabled');
    cy.findByRole('button', {name: 'Close sidebar'}).click();

    cy.get('#GROUP_montage').within(() =>
      cy.findByText('Pipeline').click({force: true}),
    );
    cy.findByTestId('DeletePipelineButton__link').click();
    cy.findByTestId('ModalFooter__confirm').click();
    cy.findByTestId('ModalFooter__confirm').should('not.exist');
    cy.get('#GROUP_montage').should('not.exist');
    cy.url().should('not.contain', '/pipelines');

    cy.findByText('images').click({force: true});
    cy.findByTestId('DeleteRepoButton__link').should('be.disabled');
    cy.findByRole('button', {name: 'Close sidebar'}).click();

    cy.get('#GROUP_edges').within(() => cy.findByText('Pipeline').click());
    cy.findByTestId('DeletePipelineButton__link').click();
    cy.findByTestId('ModalFooter__confirm').click();
    cy.findByTestId('ModalFooter__confirm').should('not.exist');
    cy.get('#GROUP_edges').should('not.exist');
    cy.url().should('not.contain', '/pipelines');

    cy.findByText('images').click({force: true});
    cy.findByTestId('DeleteRepoButton__link').should('not.be.disabled');
  });
});
