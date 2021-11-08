describe('Project', () => {
  beforeEach(() => {
    cy.login();
  });

  afterEach(() => {
    cy.logout();
  });

  it('should navigate to the project page', () => {
    cy.findAllByText('View Project').eq(0).click();
    cy.findByText('Show Jobs');
    cy.findByText('Reset Canvas');
    cy.findByText('images');
    cy.findAllByText('edges').should('have.length', 2);
  });
});
