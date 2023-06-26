describe('Dag', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.setupProject('error-opencv').visit('/');
  });
  beforeEach(() => {
    cy.visit('/');
    cy.findByRole('button', {name: /View project default/i}).click();
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
    cy.findByText('Jobs', {timeout: 10000}).click();
    cy.findByLabelText('expand filters').click();
    cy.findAllByText('Failed').filter(':visible').click();
    cy.findAllByTestId('RunsList__row', {timeout: 60000}).should(
      'have.length',
      2,
    );
    cy.findAllByTestId('RunsList__row')
      .eq(1)
      .findByTestId('DropdownButton__button')
      .click();
    cy.findAllByText('Apply Global ID and view in DAG').eq(1).click();

    cy.findByText('DAG').click();
    cy.get('#GROUP_images', {timeout: 10000}).should('exist');
    cy.get('#GROUP_edges').should('exist');
    cy.get('#GROUP_montage', {timeout: 10000}).should('not.exist');

    cy.findByTestId('Node__state-ERROR', {timeout: 12000}).should('exist');
  });

  it('should derive the correct selected repo from the url', () => {
    cy.visit('/lineage/default/repos/images');
    cy.get('#GROUP_images', {timeout: 10000}).should('be.visible');
    cy.findByTestId('Title__name').should('have.text', 'images');
  });

  it('should derive the correct selected pipeline from the url', () => {
    cy.visit('/lineage/default/pipelines/edges');
    cy.get('#GROUP_edges', {timeout: 10000}).should('be.visible');
    cy.findByTestId('Title__name').should('have.text', 'edges');
  });

  it('should update the url correctly when selecting a repo', () => {
    cy.get('#GROUP_images', {timeout: 10000}).click();
    cy.url().should('contain', '/lineage/default/repos/images');
  });

  it('should update the url correctly when selecting a pipeline', () => {
    cy.get('#GROUP_edges', {timeout: 10000}).within(() =>
      cy.findByText('Pipeline').click(),
    );
    cy.url().should('contain', '/lineage/default/pipelines/edges');
  });

  it('should update the url correctly when selecting an output repo', () => {
    cy.get('#GROUP_edges', {timeout: 10000}).within(() =>
      cy.findByText('Output').click(),
    );
    cy.url().should('contain', '/lineage/default/repos/edges');
  });

  it('should update the url correctly when selecting a status icon', () => {
    cy.get('#GROUP_edges', {timeout: 10000}).within(() =>
      cy.findByTestId('Node__state-ERROR').click(),
    );
    cy.url().should(
      'contain',
      '/lineage/default/pipelines/edges/logs?prevPath=%2Flineage%2Fdefault',
    );
  });

  it('should not update the url when selecting an egress node', () => {
    cy.get('#GROUP_', {timeout: 10000}).click({force: true});

    cy.url().should('equal', 'http://localhost:4000/lineage/default');
  });

  it('should correctly reset the DAG when DAG nodes are deleted', () => {
    cy.get('#GROUP_montage', {timeout: 10000}).within(() =>
      cy.findByText('Pipeline').click(),
    );
    cy.findByTestId('DeletePipelineButton__link').click();
    cy.findByTestId('ModalFooter__confirm').click();

    cy.url().should('not.include', 'montage');
  });
});
