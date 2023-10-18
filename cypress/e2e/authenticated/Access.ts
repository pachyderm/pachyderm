before(() => {
  cy.exec('echo "pizza" | pachctl auth use-auth-token');
  cy.deleteReposAndPipelines();
  cy.setupProject()
    .exec('pachctl auth set cluster projectViewer user:kilgore@kilgore.trout')
    .exec('pachctl auth set repo images repoReader user:kilgore@kilgore.trout')
    .logout();
});

beforeEach(() => {
  cy.login();
});

afterEach(() => {
  cy.logout();
});

after(() => {
  cy.exec('echo "pizza" | pachctl auth use-auth-token');
  cy.deleteReposAndPipelines();
});

describe('Access', () => {
  beforeEach(() => {
    cy.findAllByText(/^View(\sProject)*$/)
      .eq(0)
      .click();
  });

  it('should let non-admins see the DAG', () => {
    cy.findAllByText('edges', {timeout: 16000})
      .should('have.length', 1)
      .first()
      .click();
    cy.url().should('include', 'edges');
    cy.findByText("You don't have permission to view this pipeline").should(
      'exist',
    );

    cy.findByText('images').should('exist').click();
    cy.url().should('include', 'images');
    cy.findByText('Most Recent Commit ID').should('exist');
  });

  it('should not allow users to view repos they do not have access for', () => {
    cy.findByText('Repositories', {timeout: 30000}).click();
    cy.findByText('edges').click();
    cy.get('Detailed info').should('not.exist');
  });
});

describe('Header', () => {
  it('when in Enterprise Edition the header shows the correct app name', () => {
    cy.findByRole('banner').findByRole('heading', {
      name: 'HPE ML Data Management',
    });
  });
});
