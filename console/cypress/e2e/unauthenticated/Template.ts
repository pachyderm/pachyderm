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

    cy.exec(
      'kubectl create secret generic snowflake-connection --from-literal snow-account=xyz --from-literal snow-user=user --from-literal  snow-password=xyz',
    );
    cy.exec(
      'kubectl create secret generic huggingface-token --from-literal HF_HOME=xyz',
    );
  });

  after(() => {
    cy.deleteProjectReposAndPipelines('templates');
    cy.exec('kubectl delete secret snowflake-connection --ignore-not-found');
    cy.exec('kubectl delete secret huggingface-token --ignore-not-found');
  });

  it('should load example snowflake template and create pipeline', () => {
    cy.visit('/lineage/templates');
    cy.findByText(/create/i).click();
    cy.findByText(/pipeline from template/i).click();

    cy.findByRole('heading', {
      name: /snowflake integration/i,
    }).click();
    cy.findByText(/continue/i).click();

    cy.findByRole('textbox', {
      name: /name/i,
    }).type('snowflake_pipeline');

    cy.findByRole('textbox', {
      name: /query/i,
    }).type('xyz');

    cy.findByText(/create pipeline/i).click();
    cy.findByText('snowflake_pipeline_tick');
    cy.findByText('snowflake_pipeline');
  });

  it('should load example hugging face template and create pipeline', () => {
    cy.visit('/lineage/templates');
    cy.findByText(/create/i).click();
    cy.findByText(/pipeline from template/i).click();

    cy.findByRole('heading', {
      name: /hugging face downloader/i,
    }).click();
    cy.findByText(/continue/i).click();

    cy.findByRole('textbox', {
      name: 'name',
    }).type('hugging_face');

    cy.findByRole('textbox', {
      name: 'hf_name',
    }).type('xyz');

    cy.findByText(/create pipeline/i).click();
    cy.findByText('hugging_face_cron');
    cy.findByText('hugging_face');
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
