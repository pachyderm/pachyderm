describe('Docker Build', () => {
  it('should load correctly and render a header', () => {
    cy.visit('/');
    cy.findByRole('heading', {name: 'default'});
  });

  it('should proxy request', () => {
    cy.multiLineExec(
      `pachctl create repo images
      echo "hello, this is a text file" | pachctl put file images@master:data.txt -f -
      `,
    ).visit('/lineage/default/repos/images/latest');
    cy.findByText('data.txt').click();

    cy.findByText('hello, this is a text file').should('exist');
  });
});
