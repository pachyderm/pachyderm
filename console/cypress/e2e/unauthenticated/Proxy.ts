describe('Proxy', () => {
  beforeEach(() => {
    cy.deleteReposAndPipelines();
    cy.visit('/');
  });

  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('should proxy request', () => {
    cy.multiLineExec(
      `pachctl create repo images
      echo "hello, this is a text file" | pachctl put file images@master:data.txt -f -
      `,
    ).visit('/lineage/default/repos/images/latest');
    cy.findAllByText('data.txt').last().click();

    cy.findByText('hello, this is a text file').should('exist');
  });
});
