describe('Pipeline steps', () => {
  before(() => {
    cy.setupProject('error-opencv').visit('/');
  })

  beforeEach(() => {
    cy.findAllByText(/^View(\sProject)*$/).eq(0).click();
  });

  afterEach(() => {
    cy.visit('/')
  })

  after(() => {
    cy.deleteReposAndPipelines();
  })

  it('should allow a user to select a subset of pipelines to inspect jobs and apply a global ID', () => {
    cy.findByText('Pipeline Steps').click();
    cy.findAllByTestId('PipelineStepsList__row').should('have.length', 2)

    cy.findAllByTestId('PipelineStepsList__row').eq(1).click();
    cy.findByText('Detailed info for 1 pipeline step');
    cy.findByRole('tab', {name: 'Jobs'}).click();
    
    cy.findAllByTestId('JobsList__row').should('have.length', 1)

    cy.findAllByTestId('JobsList__row').first().within(() => cy.findByTestId('DropdownButton__button').click());
    cy.findByText('Apply Global ID and view in DAG').click()

    cy.get('#GROUP_images').should('exist');
    cy.get('#GROUP_edges').should('exist');
    cy.get('#GROUP_montage').should('not.exist');
  })
});
