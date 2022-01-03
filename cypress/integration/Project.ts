describe('Project', () => {
  beforeEach(() => {
    cy.login().then(() => {
      cy.findAllByText('View Project').eq(0).click();
    });
  });

  afterEach(() => {
    cy.logout();
  });

  it('should navigate to the project page', () => {
    cy.findByText('Show Jobs');
    cy.findByText('Reset Canvas');
    cy.findByText('images');
    cy.findAllByText('edges').should('have.length', 2);
  });

  it('should show the correct number of commits', () => {

    cy.findByText('images').click();
    cy.findAllByText('Committed', {exact: false}).should('have.length', 2);
    cy.findAllByText('(57.27 KB)', {exact: false}).should('have.length', 2);
    cy.findByRole('button', {name: 'Close'}).click();
    cy.findAllByText('edges').eq(0).click();
    cy.findAllByText('Committed', {exact: false}).should('have.length', 1);
    cy.findByText('(22.22 KB)', {exact: false});

  })
});
