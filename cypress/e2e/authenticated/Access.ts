before(() => {
  cy.exec('echo "pizza" | pachctl auth use-auth-token');
  cy.deleteReposAndPipelines();
  cy.setupProject()
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
    cy.visit('/');
    cy.findAllByText(/^View(\sProject)*$/)
      .eq(0)
      .click();
  });

  it('should let non-admins see the DAG', () => {
    const edgeNodes = cy.findAllByText('edges', {timeout: 16000});
    edgeNodes.should('have.length', 1);
    edgeNodes.first().click();
    cy.url().should('not.include', 'edges');
    const imagesNode = cy.findByText('images');
    imagesNode.should('exist');
    imagesNode.click();
    cy.url().should('include', 'images');
  });

  it('should not allow users to view repos they do not have access for', () => {
    cy.findByText('Repositories', {timeout: 30000}).click();
    const edges = cy.findByText('edges');
    edges.click();
    cy.get('Detailed info').should('not.exist');
  });
});

describe('Header', () => {
  it('when in Enterprise Edition the header shows the correct app name', () => {
    cy.visit('/');

    cy.findByRole('banner').findByRole('heading', {
      name: 'HPE ML Data Management',
    });
  });
});
