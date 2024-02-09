describe('Dag', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.setupProject('error-opencv-with-one-success');
  });
  beforeEach(() => {
    cy.visit('/lineage/default');
  });
  after(() => {
    cy.deleteReposAndPipelines();
  });

  it('simple date filter and jq filters work correctly', () => {
    // wait for dag to load
    cy.findByRole('button', {
      name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 repo',
      timeout: 10000,
    }).should('exist');

    // open job
    cy.findByRole('button', {name: /previous jobs/i}).click();
    cy.findAllByTestId('GlobalFilter__status_passing', {
      timeout: 30_000,
    }).should('have.length', 1);
    cy.findAllByTestId('GlobalFilter__status_failing', {
      timeout: 30_000,
    }).should('have.length', 2);

    // the simple date filter works
    cy.findByText(/\(2\) last hour/i).click();
    cy.findAllByTestId('SearchResultItem__container').should('have.length', 2);
    cy.findByRole('button', {name: 'Clear'}).click();
    cy.findAllByTestId('SearchResultItem__container').should('have.length', 3);

    // ensure the right items are shown in the global id filter on the dag
    cy.findAllByRole('listitem').eq(0).click();

    cy.findByRole('button', {
      name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 repo',
      timeout: 10000,
    }).should('exist');
    //edges
    cy.findByRole('button', {
      name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
      timeout: 10000,
    }).should('exist');
    //montage
    cy.findByRole('button', {
      name: 'GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6 pipeline',
      timeout: 10000,
    }).should('exist');

    // ensure the right items are shown when you click a global ID
    cy.findAllByRole('listitem').eq(1).click();

    cy.findByRole('button', {
      name: 'GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65 repo',
      timeout: 10000,
    }).should('exist');
    //edges
    cy.findByRole('button', {
      name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
      timeout: 10000,
    }).should('exist');
    //montage
    cy.findByRole('button', {
      name: 'GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6 pipeline',
      timeout: 10000,
    }).should('not.exist');
  });
});
