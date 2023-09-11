describe('Pipelines', () => {
  before(() => {
    cy.setupProject('error-opencv').visit('/');
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

  it('should allow a user to select a subset of pipelines to inspect jobs and apply a global ID', () => {
    cy.get('#GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65').should('exist');
    cy.get('#GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838').should('exist');
    cy.get('#GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6').should('exist');

    cy.findByText('Pipelines').click();
    cy.findAllByTestId('PipelineListRow__row').should('have.length', 2);

    cy.findAllByTestId('PipelineListRow__row').eq(1).click();
    cy.findByText('Detailed info for 1 pipeline');
    cy.findByRole('tab', {name: 'Jobs'}).click();

    cy.findAllByTestId('JobsList__row').should('have.length', 1);

    cy.findAllByTestId('JobsList__row')
      .first()
      .within(() => cy.findByTestId('DropdownButton__button').click());
    cy.findByText('Apply Global ID and view in DAG').click();

    cy.get('#GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65').should('exist');
    cy.get('#GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838').should('exist');
    cy.get('#GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6').should('not.exist');
  });
});
