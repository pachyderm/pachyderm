describe('Docker Build', () => {
  it('should load correctly and render a header', () => {
    cy.visit('/');
    cy.findByRole('heading', {name: 'default'});
  });
});
