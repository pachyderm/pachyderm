describe('Access', () => {
  before(() => {
    cy.authenticatePachctl()
      .setupProject()
      .exec('pachctl auth set repo images repoReader user:john-doe@pachyderm.io')
      .logout()
      .login('john-doe@pachyderm.io')
  });

  after(() => {
    cy.deleteReposAndPipelines();
  })

  afterEach(() => {
    cy.visit('/')
  })

  describe('Lineage View', () => {
    it('should let non-admins see the DAG', () => {
      cy.findByText('Skip tutorial').click();
      cy.findAllByText(/^View(\sProject)*$/).eq(0).click();
      const edgeNodes = cy.findAllByText('edges', {timeout: 16000});
      edgeNodes.should('have.length', 1);
      edgeNodes.first().click();
      cy.url().should('not.include', 'edges');
      const imagesNode = cy.findByText('images');
      imagesNode.should('exist');
      imagesNode.click();
      cy.url().should('include', 'images');
    });
  })

  describe('List View', () => {
    beforeEach(() => {
      cy.findByText('Skip tutorial', {timeout: 8000}).click();
      cy.findAllByText(/^View(\sProject)*$/).eq(0).click();
      cy.findByText('View List').click();
    })
    it('should select the first repo the user has access to', () => {
      cy.url().should('include', 'images');
    })

    it('should not allow users to view repos they do not have access for', () => {
      const edges = cy.findByText('edges', {timeout: 16000});
      edges.click();
      cy.url().should('not.include', 'edges');
    })
  })
});
