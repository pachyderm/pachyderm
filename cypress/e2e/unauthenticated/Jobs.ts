describe('Jobs', () => {
    before(() => {
      cy.setupProject().visit('/');
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
  
    it('should show jobset details', () => {
      cy.findByText('Jobs', {timeout: 12000}).click();
      cy.findByText('See Details').click();
      cy.findByTestId('InfoPanel__commitLink', {timeout: 12000}).should('contain', /^[a-zA-Z0-9_-]{32}/);
      cy.findByTestId('InfoPanel__pipeline', {timeout: 12000}).should('have.text', 'edges');
      // wait for job to finish
      cy.findByTestId('InfoPanel__state', {timeout: 12000}).should('have.text', 'Success');
      cy.findByTestId('RuntimeStats__started').should('include.text', 'ago');
      cy.findByTestId('RuntimeStats__duration').invoke('text').should('match', /^\s?\d+\sseconds?$/);
      cy.findByTestId('InfoPanel__success').should('have.text', '1');
      cy.findByTestId('InfoPanel__skipped').should('have.text', '0');
      cy.findByTestId('InfoPanel__failed').should('have.text', '0');
      cy.findByTestId('InfoPanel__recovered').should('have.text', '0');
      cy.findByTestId('InfoPanel__total').should('have.text', '1');
    });
  });
