describe('Access', () => {
  before(() => {
    cy.setupProject()
      .exec('pachctl auth set repo images repoReader user:john-doe@pachyderm.io')
      .logout()
      .login('john-doe@pachyderm.io')
  });

  after(() => {
    cy.deleteReposAndPipelines().logout();
  })

  it('should let non-admins see the DAG', () => {
    cy.findAllByText(/^View(\sProject)*$/).eq(0).click();
    const edgeNodes = cy.findAllByText('edges', {timeout: 16000});
    edgeNodes.should('have.length', 2);
    edgeNodes.first().click();
    cy.url().should('not.include', 'edges');
    const imagesNode = cy.findByText('images');
    imagesNode.should('exist');
    imagesNode.click();
    cy.url().should('include', 'images');
  })
});
