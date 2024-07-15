describe('Pipeline Templates', () => {
  before(() => {
    // We can self host the template file by committing it to a repo
    cy.multiLineExec(
      `
    pachctl create project templates
    pachctl create repo images --project templates
    pachctl put file images@master:edges.jsonnet -f cypress/fixtures/edges.jsonnet  --project templates
    `,
    ).visit('/');
  });

  after(() => {
    cy.deleteProjectReposAndPipelines('templates');
  });

  it('should create simple pipeline from template', () => {
    cy.visit('/lineage/templates');

    cy.findByText(/create/i).click();
    cy.findByText(/pipeline from template/i).click();

    //Retrieve the file using the pfs api directly
    cy.findByRole('textbox').type(
      `${Cypress.config(
        'baseUrl',
      )}/proxyForward/pfs/templates/images/master/edges.jsonnet`,
    );

    cy.findByText(/template uploaded/i);
    cy.findByText(/continue/i).click();
    cy.findByText(/create pipeline/i).click();
    cy.findByText(/edges-1/i);
  });
});
