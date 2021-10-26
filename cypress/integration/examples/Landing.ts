describe('Landing', () => {
  beforeEach(() => {
    cy.login();
  });

  afterEach(() => {
    cy.logout();
  });

  it('should show deafult project info', () => {
    cy.findByText('Default Pachyderm project.')
  });

  it('should navigate to the project page', () => {
    cy.findAllByText('View Project').eq(0).click();
    cy.findByText('Show Jobs');
    cy.findByText('Reset Canvas');
  });
});
