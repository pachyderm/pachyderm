describe('Dag', () => {
  before(() => {
    cy.setupProject('error-opencv').visit('/');
  });
  beforeEach(() => {
    cy.findByText('default')
      .parent()
      .findByRole('button', {name: /View/i})
      .click();
  });
  afterEach(() => {
    cy.visit('/');
  });
  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('should render the entire dag', () => {
    cy.get('#GROUP_images', {timeout: 10000})
      .should('exist')
      .findByText('images')
      .should('exist');

    cy.get('#GROUP_edges')
      .should('exist')
      .findAllByText('edges')
      .should('exist');

    cy.get('#GROUP_montage')
      .should('exist')
      .findAllByText('montage')
      .should('exist');

    cy.get('#GROUP_').should('exist');
  });

  it('should render a sub-dag with a globalId filter', () => {
    cy.findByRole('link', {
      name: /jobs/i,
      timeout: 10000,
    }).click();

    cy.findAllByText('Failure', {timeout: 60000})
      .should('exist')
      .last()
      .click({force: true});

    cy.findByTestId('CommitIdCopy__id').invoke('text').then((jobId) => {
      cy.findByText('Filter by Global ID').click();
      cy.findByTestId('GlobalFilter__name').type(jobId);
      cy.findByText('Apply').click();
    });

    cy.get("#GROUP_images", { timeout: 10000 }).should("exist");
    cy.get("#GROUP_edges").should("exist");
    cy.get('#GROUP_montage', { timeout: 10000 }).should("not.exist");

    cy.findByTestId('Node__state-ERROR', {timeout: 12000}).should('exist');
  });

  it('should derive the correct selected repo from the url', () => {
    cy.visit('/lineage/default/repos/images/branch/default');
    const imageNode = cy.get("#GROUP_images", { timeout: 10000 });
    imageNode.should('be.visible');
    cy.findByTestId("Title__name").should("have.text", "images");
  });

  it('should derive the correct selected pipeline from the url', () => {
    cy.visit('/lineage/default/pipelines/edges');
    const edgesNode = cy.get("#GROUP_edges", { timeout: 10000 });
    edgesNode.should('be.visible');
    cy.findByTestId("Title__name").should("have.text", "edges");
  });

  it('should update the url correctly when selecting a repo', () => {
    const imageNode = cy.get("#GROUP_images", { timeout: 10000 });
    imageNode.click();
    cy.url().should("contain", "/lineage/default/repos/images/branch/default");
  });

  it('should update the url correctly when selecting a pipeline', () => {
    const edgesPipelineNode = cy.get("#GROUP_edges", { timeout: 10000 });
    edgesPipelineNode.within(() => cy.findByText('Pipeline').click());
    cy.url().should("contain", "/lineage/default/pipelines/edges");
  });

  it('should update the url correctly when selecting an output repo', () => {
    const edgesPipelineNode = cy.get("#GROUP_edges", { timeout: 10000 });
    edgesPipelineNode.within(() => cy.findByText('Output').click());
    cy.url().should("contain", "/lineage/default/repos/edges/branch/default");
  });

  it('should not update the url when selecting an egress node', () => {
    const egressNode = cy.get("#GROUP_", { timeout: 10000 });
    egressNode.click({ force: true });

    cy.url().should('equal', "http://localhost:4000/lineage/default");
  });

  it('should correctly reset the DAG when DAG nodes are deleted', () => {
    const montagePipelineNode = cy.get("#GROUP_montage", { timeout: 10000 });
    montagePipelineNode.within(() => cy.findByText('Pipeline').click());
    cy.findByTestId('DeletePipelineButton__link').click();
    cy.findByTestId('ModalFooter__confirm').click();

    cy.url().should('not.include', 'montage');
  });
})
