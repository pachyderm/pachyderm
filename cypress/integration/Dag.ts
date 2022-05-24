describe('Dag', () => {
  before(() => {
    cy.setupProject('error-opencv').visit('/');
  });
  beforeEach(() => {
    cy.findByText('Skip tutorial').click();
    cy.findAllByText(/^View(\sProject)*$/).eq(0).click();
  });
  afterEach(() => {
    cy.visit('/')
  });
  after(() => {
    cy.deleteReposAndPipelines().logout();
  });

  it('should render the entire dag', () => {
    const imageNode = cy.get("#GROUP_images_repo", { timeout: 10000 });
    imageNode.should("exist");
    imageNode.findByText("images").should("exist");
    const edgesRepoNode = cy.get("#GROUP_edges_repo");
    edgesRepoNode.should("exist");
    edgesRepoNode.findAllByText("edges").should("exist");
    const edgesPipelineNode = cy.get('#GROUP_edges');
    edgesPipelineNode.should("exist");
    edgesPipelineNode.findAllByText("edges").should("exist");
    const montageRepoNode = cy.get('#GROUP_montage_repo')
    montageRepoNode.should('exist');
    montageRepoNode.findAllByText("montage").should("exist");
    const montagePipelineNode = cy.get('#GROUP_montage');
    montagePipelineNode.should('exist');
    montagePipelineNode.findAllByText("montage").should("exist");
    cy.get("#GROUP_").should("exist");
  });

  it('should render a sub-dag with a globalId filter', () => {
    const jobs = cy.findByTestId("ProjectSideNav__seeJobs", {timeout: 10000});
    jobs.click();

    const failedJob = cy.findAllByText("Failure", {timeout: 60000});
    failedJob.should('exist');
    failedJob.last().click({force: true});

    cy.findByTestId('CommitIdCopy__id').invoke('text').then((jobId) => {
      cy.findByText('Filter by Global ID').click();
      cy.findByTestId('GlobalFilter__name').type(jobId);
      cy.findByText('Apply').click();
    });

    cy.get("#GROUP_images_repo", { timeout: 10000 }).should("exist");
    cy.get("#GROUP_edges_repo").should("exist");
    cy.get('#GROUP_montage_repo').should("not.exist");

    cy.findByTestId('Node__state-ERROR', {timeout: 12000}).should('exist');
  });

  it('should derive the correct selected repo from the url', () => {
    cy.visit('/lineage/default/repos/images/branch/master');
    const imageNode = cy.get("#GROUP_images_repo", { timeout: 10000 });
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
    const imageNode = cy.get("#GROUP_images_repo", { timeout: 10000 });
    imageNode.click();
    cy.url().should("contain", "/lineage/default/repos/images/branch/master");
  });

  it('should update the url correctly when selecting a pipeline', () => {
    const edgesPipelineNode = cy.get("#GROUP_edges", { timeout: 10000 });
    edgesPipelineNode.click();
    cy.url().should("contain", "/lineage/default/pipelines/edges");
  });

  it('should not update the url when selecting an egress node', () => {
    const egressNode = cy.get("#GROUP_", { timeout: 10000 });
    egressNode.click({ force: true });

    cy.url().should('equal', "http://localhost:4000/lineage/default");
  });

  it('should correctly reset the DAG when DAG nodes are deleted', () => {
    cy.get("#GROUP_montage", { timeout: 10000 }).click();
    cy.findByTestId('DeletePipelineButton__link').click();
    cy.findByTestId('ModalFooter__confirm').click();

    cy.url().should('not.include', 'montage');
  });
})
