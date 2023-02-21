describe('Access', () => {
  before(() => {
    cy.authenticatePachctl()
      .setupProject()
      .exec(
        'pachctl auth set repo images repoReader user:john-doe@pachyderm.io',
      )
      .logout()
      .login('john-doe@pachyderm.io');
  });

  beforeEach(() => {
    cy.findAllByText(/^View(\sProject)*$/)
      .eq(0)
      .click();
  });

  after(() => {
    cy.deleteReposAndPipelines();
  });

  afterEach(() => {
    cy.visit('/');
  });

  describe('Access', () => {
    it('should let non-admins see the DAG', () => {
      const edgeNodes = cy.findAllByText('edges', {timeout: 16000});
      edgeNodes.should('have.length', 1);
      edgeNodes.first().click();
      cy.url().should('not.include', 'edges');
      const imagesNode = cy.findByText('images');
      imagesNode.should('exist');
      imagesNode.click();
      cy.url().should('include', 'images');
    });

    it('should not allow users to view repos they do not have access for', () => {
      cy.findByText('Repositories', {timeout: 30000}).click();
      const edges = cy.findByText('edges');
      edges.click();
      cy.get('Detailed info').should("not.exist");
    });
  });
});
