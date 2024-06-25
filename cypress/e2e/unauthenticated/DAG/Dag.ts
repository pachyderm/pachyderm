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
    // images node
    cy.findByRole('button', {
      name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 repo',
      timeout: 10000,
    }).should('exist');
    // images -> edges
    cy.get(
      'path[id="8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65_d0e1e9a51269508c3f11c0e64c721c3ea6204838"]',
    ).should('exist');
    // images -> montage
    cy.get(
      'path[id="8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6"]',
    ).should('exist');

    //edges node
    cy.findByRole('button', {
      name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
      timeout: 10000,
    }).should('exist');
    // edges -> montage
    cy.get(
      'path[id="d0e1e9a51269508c3f11c0e64c721c3ea6204838_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6"]',
    ).should('exist');

    //montage node
    cy.findByRole('button', {
      name: 'GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6 pipeline',
      timeout: 10000,
    }).should('exist');

    cy.findByText('s3://test').should('exist');
  });

  it('should derive the correct selected repo from the url', () => {
    cy.visit('/lineage/default/repos/images');
    cy.findByRole('button', {
      name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 repo',
      timeout: 10000,
    }).should('be.visible');
    cy.findByTestId('Title__name').should('have.text', 'images');
  });

  it('should derive the correct selected pipeline from the url', () => {
    cy.visit('/lineage/default/pipelines/edges');
    cy.findByRole('button', {
      name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
      timeout: 10000,
    }).should('be.visible');
    cy.findByTestId('Title__name').should('have.text', 'edges');
  });

  it('should update the url correctly when selecting a repo', () => {
    cy.findByRole('button', {
      name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 repo',
      timeout: 10000,
    }).click();
    cy.url().should('contain', '/lineage/default/repos/images');
  });

  it('should update the url correctly when selecting a pipeline', () => {
    cy.findByRole('button', {
      name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
      timeout: 10000,
    }).click();
    cy.url().should('contain', '/lineage/default/pipelines/edges');
  });

  it('should update the url correctly when selecting an output repo', () => {
    cy.findByRole('button', {
      name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 repo',
      timeout: 10000,
    }).click();
    cy.url().should('contain', '/lineage/default/repos/edges');
  });

  it('should update the url correctly when selecting a status icon', () => {
    cy.findByRole('button', {
      name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 logs',
      timeout: 10000,
    }).click();
    cy.url().should(
      'contain',
      '/lineage/default/pipelines/edges/logs?prevPath=%2Flineage%2Fdefault',
    );
  });

  it('should correctly reset the DAG when DAG nodes are deleted', () => {
    cy.findByRole('button', {
      name: 'GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6 pipeline',
      timeout: 10000,
    }).click();

    cy.findByRole('button', {
      name: /pipeline-actions-menu/i,
    }).click();
    cy.findByRole('menuitem', {
      name: /delete pipeline/i,
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

  it('should update the url correctly when selecting a to upload a file', () => {
    cy.findByRole('button', {
      name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 file upload',
      timeout: 10000,
    }).click();
    cy.url().should('contain', '/lineage/default/repos/images/upload');
  });

  it('should show correct hover stats when pipeline is hovered', () => {
    cy.findByRole('button', {
      name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
    })
      .should('exist')
      .trigger('mouseover');

    cy.findByText('Workers active', {timeout: 10000});
    cy.findByText('1/1', {timeout: 10000});
  });

  it('should show correct hover stats when edge is hovered', () => {
    // images -> edges
    cy.get(
      'path[id="8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65_d0e1e9a51269508c3f11c0e64c721c3ea6204838"]',
    )
      .should('exist')
      .trigger('mouseover');

    cy.findByTestId('HoverStats__link').should('contain.text', 'images edges', {
      timeout: 10000,
    });
  });
});
