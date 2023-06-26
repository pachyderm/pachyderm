describe('Dag', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.setupProject('error-opencv');
  });
  beforeEach(() => {
    cy.visit('/lineage/default');
  });
  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('should render the entire dag', () => {
    cy.findByRole('button', {
      name: 'GROUP_images repo',
      timeout: 10000,
    }).should('exist');

    cy.findByRole('button', {
      name: 'GROUP_edges pipeline',
      timeout: 10000,
    }).should('exist');

    cy.findByRole('button', {
      name: 'GROUP_montage pipeline',
      timeout: 10000,
    }).should('exist');

    cy.findByRole('button', {
      name: 'GROUP_ egress',
    }).should('exist');
  });

  it('should render a sub-dag with a globalId filter', () => {
    cy.findByText('Jobs', {timeout: 10000}).click();
    cy.findByLabelText('expand filters').click();
    cy.findAllByText('Failed').filter(':visible').click();

    cy.findByRole('rowgroup', {
      name: 'runs list',
      timeout: 60000,
    }).within(() => {
      cy.findAllByRole('row').should('have.length', 2);
    });

    cy.findAllByTestId('RunsList__row')
      .eq(1)
      .findByTestId('DropdownButton__button')
      .click();
    cy.findByRole('menuitem', {
      name: /apply global id and view in dag/i,
    }).click();

    cy.findByRole('button', {
      name: 'GROUP_images repo',
      timeout: 10000,
    }).should('exist');
    cy.findByRole('button', {
      name: 'GROUP_edges pipeline',
    }).should('exist');
    cy.findByRole('button', {
      name: 'GROUP_montage pipeline',
      timeout: 10000,
    }).should('not.exist');

    cy.findByTestId('Node__state-ERROR', {timeout: 12000}).should('exist');
  });

  it('should derive the correct selected repo from the url', () => {
    cy.visit('/lineage/default/repos/images');
    cy.findByRole('button', {
      name: 'GROUP_images repo',
      timeout: 10000,
    }).should('be.visible');
    cy.findByTestId('Title__name').should('have.text', 'images');
  });

  it('should derive the correct selected pipeline from the url', () => {
    cy.visit('/lineage/default/pipelines/edges');
    cy.findByRole('button', {
      name: 'GROUP_edges pipeline',
      timeout: 10000,
    }).should('be.visible');
    cy.findByTestId('Title__name').should('have.text', 'edges');
  });

  it('should update the url correctly when selecting a repo', () => {
    cy.findByRole('button', {
      name: 'GROUP_images repo',
      timeout: 10000,
    }).click();
    cy.url().should('contain', '/lineage/default/repos/images');
  });

  it('should update the url correctly when selecting a pipeline', () => {
    cy.findByRole('button', {
      name: 'GROUP_edges pipeline',
      timeout: 10000,
    }).click();
    cy.url().should('contain', '/lineage/default/pipelines/edges');
  });

  it('should update the url correctly when selecting an output repo', () => {
    cy.findByRole('button', {
      name: 'GROUP_edges repo',
      timeout: 10000,
    }).click();
    cy.url().should('contain', '/lineage/default/repos/edges');
  });

  it('should update the url correctly when selecting a status icon', () => {
    cy.findByRole('button', {
      name: 'GROUP_edges logs',
      timeout: 10000,
    }).click();
    cy.url().should(
      'contain',
      '/lineage/default/pipelines/edges/logs?prevPath=%2Flineage%2Fdefault',
    );
  });

  it('should not update the url when selecting an egress node', () => {
    cy.findByRole('button', {
      name: 'GROUP_ egress',
    }).click({force: true});

    cy.url().should('equal', 'http://localhost:4000/lineage/default');
  });

  it('should correctly reset the DAG when DAG nodes are deleted', () => {
    cy.findByRole('button', {
      name: 'GROUP_montage pipeline',
      timeout: 10000,
    }).click();

    cy.findByRole('button', {
      name: /delete/i,
    }).click();
    cy.findByRole('dialog').within(() =>
      cy
        .findByRole('button', {
          name: /delete/i,
        })
        .click(),
    );

    cy.url().should('not.include', 'montage');
  });
});
