describe('JobSets', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.setupProject('error-opencv').visit('/');
  });

  beforeEach(() => {
    cy.visit('/');
    cy.findAllByText(/^View(\sProject)*$/)
      .eq(0)
      .click();
  });

  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('should allow a user to select a subset of jobsets to inspect subjobs', () => {
    cy.findByText('Jobs').click();
    cy.findByLabelText('expand filters').click();
    cy.findAllByText('Failed').filter(':visible').click();
    cy.findAllByTestId('RunsList__row', {timeout: 60000}).should(
      'have.length',
      2,
    );

    cy.findAllByTestId('RunsList__row').eq(0).click();
    cy.findAllByTestId('RunsList__row').eq(1).click();
    cy.findByText('Detailed info for 2 jobs');
    cy.findByText('Subjobs').click();

    cy.findAllByTestId('JobsList__row').should('have.length', 2);

    cy.findAllByTestId('JobsList__row')
      .eq(1)
      .within(() => cy.findByLabelText('Failed datums').click());

    cy.findByTestId('Filter__FAILEDChip').should('exist');
    cy.findByTestId('SidePanel__closeLeft').click();
    cy.findAllByTestId('LogRow__base')
      .first()
      .parent()
      .parent()
      .scrollTo('bottom');
    cy.findByText(
      /AttributeError: 'NoneType' object has no attribute 'shape'/,
    ).should('be.visible');
  });
});
