describe('Header', () => {
  it('when in Community Edition it should show the correct app name', () => {
    cy.visit('/');
    cy.findByRole('banner').findByRole('heading', {name: 'Console'});
  });
});
