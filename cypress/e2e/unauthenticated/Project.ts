describe('Project', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.multiLineExec(
      `
    pachctl create repo images
    echo '{"pipeline": {"name": "edges"} ,"transform": {"cmd": ["sh"],"stdin": ["sleep 0"]},"input": {"pfs": {"glob": "/*","repo": "images"}}}'  | pachctl create pipeline
    echo '{"pipeline": {"name": "montage"} ,"transform": {"cmd": ["sh"],"stdin": ["sleep 0"]},  "input": {"cross": [{"pfs": {"glob": "/","repo": "images"}}, {"pfs": {"glob": "/","repo": "edges"}}]},}'  | pachctl create pipeline
  `,
    ).visit('/');
  });

  beforeEach(() => {
    cy.findAllByText(/^View(\sProject)*$/, {timeout: 30_000})
      .eq(0)
      .click();
  });

  afterEach(() => {
    cy.visit('/');
  });

  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('should enable the pipeline and repo deletion buttons when all downstream pipelines and repos are deleted', () => {
    cy.findByText('images', {timeout: 30_000}).click();
    cy.findByRole('button', {name: 'Repo Actions'}).click();
    cy.findByRole('menuitem', {name: 'Delete Repo'}).should('be.disabled');
    cy.findByRole('button', {name: 'Close sidebar'}).click();

    cy.get('#GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838').within(() =>
      cy.findByText('Pipeline').click(),
    );
    cy.findByRole('button', {name: 'Pipeline Actions'}).click();
    cy.findByRole('menuitem', {name: 'Delete Pipeline'}).should('be.disabled');
    cy.findByRole('button', {name: 'Close sidebar'}).click();

    cy.get('#GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6').within(() =>
      cy.findByText('Pipeline').click({force: true}),
    );
    cy.findByRole('button', {name: 'Pipeline Actions'}).click();
    cy.findByRole('menuitem', {name: 'Delete Pipeline'}).click();
    cy.findByTestId('ModalFooter__confirm').click();
    cy.findByTestId('ModalFooter__confirm').should('not.exist');
    cy.get('#GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6').should(
      'not.exist',
    );
    cy.url().should('not.contain', '/pipelines');

    cy.findByText('images').click({force: true});
    cy.findByRole('button', {name: 'Repo Actions'}).click();
    cy.findByRole('menuitem', {name: 'Delete Repo'}).should('be.disabled');
    cy.findByRole('button', {name: 'Close sidebar'}).click();

    cy.get('#GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838').within(() =>
      cy.findByText('Pipeline').click(),
    );
    cy.findByRole('button', {name: 'Pipeline Actions'}).click();
    cy.findByRole('menuitem', {name: 'Delete Pipeline'}).click();
    cy.findByTestId('ModalFooter__confirm').click();
    cy.findByTestId('ModalFooter__confirm').should('not.exist');
    cy.get('#GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838').should(
      'not.exist',
    );
    cy.url().should('not.contain', '/pipelines');

    cy.findByText('images').click({force: true});
    cy.findByRole('button', {name: 'Repo Actions'}).click();
    cy.findByRole('menuitem', {name: 'Delete Repo'}).should('be.enabled');
  });
});
