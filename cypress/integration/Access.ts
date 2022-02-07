describe('Access', () => {
  beforeEach(() => {
    cy.login('john-doe@pachyderm.io');
  });

  afterEach(() => {
    cy.logout();
  });

  it('should let non-admins see the DAG', () => {
    cy.findAllByText('View Project').eq(0).click();
    const edgeNodes = cy.findAllByText('edges');
    edgeNodes.should('have.length', 2);
    edgeNodes.first().click();
    cy.url().should('not.include', 'edges');
    const imagesNode = cy.findByText('images');
    imagesNode.should('exist');
    imagesNode.click();
    cy.url().should('include', 'images');
  })
});
