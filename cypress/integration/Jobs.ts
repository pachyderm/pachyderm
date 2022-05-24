describe('Jobs', () => {
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
      cy.deleteReposAndPipelines().logout();
    })
  
    it('should show jobset details', () => {
      cy.findByText('Jobs', {timeout: 12000}).click();
      cy.findByText('See Details').click();
      cy.findByTestId('InfoPanel__commitLink', {timeout: 12000}).should('contain', /^[a-zA-Z0-9_-]{32}/);
      cy.findByTestId('InfoPanel__pipeline').should('have.text', 'edges');
      // wait for job to finish
      cy.findByTestId('InfoPanel__state', {timeout: 12000}).should('have.text', 'Success');
      cy.findByTestId('InfoPanel__started').should('include.text', 'ago');
      cy.findByTestId('InfoPanel__duration').invoke('text').should('match', /^\d+\sseconds?$/);
      cy.findByTestId('InfoPanel__processed').should('have.text', '1');
      cy.findByTestId('InfoPanel__skipped').should('have.text', '0');
      cy.findByTestId('InfoPanel__failed').should('have.text', '0');
      cy.findByTestId('InfoPanel__recovered').should('have.text', '0');
      cy.findByTestId('InfoPanel__total').should('have.text', '1');
    });
  });
