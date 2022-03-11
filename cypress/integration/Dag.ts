describe('Dag', () => {
  before(() => {
    cy.setupProject('error-opencv').visit('/');
  });
  beforeEach(() => {
    cy.findAllByText('View Project').eq(0).click();
  });
  afterEach(() => {
    cy.visit('/')
  });
  after(() => {
    cy.deleteReposAndPipelines().logout();
  });

  it('should render the entire dag', () => {
    const imageNode = cy.get("#images_repoGROUP", { timeout: 10000 });
    imageNode.should("exist");
    imageNode.findByText("images").should("exist");
    const edgesRepoNode = cy.get("#edges_repoGROUP");
    edgesRepoNode.should("exist");
    edgesRepoNode.findAllByText("edges").should("exist");
    const edgesPipelineNode = cy.get('#edgesGROUP');
    edgesPipelineNode.should("exist");
    edgesPipelineNode.findAllByText("edges").should("exist");
    const montageRepoNode = cy.get('#montage_repoGROUP')
    montageRepoNode.should('exist');
    montageRepoNode.findAllByText("montage").should("exist");
    const montagePipelineNode = cy.get('#montageGROUP');
    montagePipelineNode.should('exist');
    montagePipelineNode.findAllByText("montage").should("exist");
    cy.get("#GROUP").should("exist");
  });

  it('should render failed nodes', () => {
    const jobs = cy.findByTestId("ProjectSideNav__seeJobs", { timeout: 10000 });
    jobs.click();

    const failedJob = cy.findAllByText("Failure", { timeout: 120000 });
    failedJob.should('exist');
    failedJob.first().click();

    cy.findByTestId('Node__state-ERROR').should('exist');
  });

  it('should derive the correct selected repo from the url', () => {
    cy.visit('/lineage/default/repos/images/branch/master');
    const imageNode = cy.get("#images_repoGROUP", { timeout: 10000 });
    imageNode.should('be.visible');
    cy.findByTestId("Title__name").should("have.text", "images");
  });

  it('should derive the correct selected pipeline from the url', () => {
    cy.visit('/lineage/default/pipelines/edges');
    const edgesNode = cy.get("#edgesGROUP", { timeout: 10000 });
    edgesNode.should('be.visible');
    cy.findByTestId("Title__name").should("have.text", "edges");
  });

  it('should update the url correctly when selecting a repo', () => {
    const imageNode = cy.get("#images_repoGROUP", { timeout: 10000 });
    imageNode.click();
    cy.url().should("contain", "/lineage/default/repos/images/branch/master");
  });

  it('should update the url correctly when selecting a pipeline', () => {
    const edgesPipelineNode = cy.get("#edgesGROUP", { timeout: 10000 });
    edgesPipelineNode.click();
    cy.url().should("contain", "/lineage/default/pipelines/edges");
  });

  it('should not update the url when selecting an egress node', () => {
    const egressNode = cy.get("#GROUP", { timeout: 10000 });
    egressNode.click({ force: true });

    cy.url().should('equal', "http://localhost:4000/lineage/default");
  });

  it('should correctly reset the DAG when DAG nodes are deleted', () => {
    cy.get("#montageGROUP", { timeout: 10000 }).click();
    cy.findByTestId('DeletePipelineButton__link').click();
    cy.findByTestId('ModalFooter__confirm').click();

    cy.url().should('not.include', 'montage');
  });
})
