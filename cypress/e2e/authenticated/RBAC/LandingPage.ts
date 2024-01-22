beforeEach(() => {
  cy.exec('echo "pizza" | pachctl auth use-auth-token');
  cy.deleteReposAndPipelines().logout();
});

after(() => {
  cy.visit('/');
  cy.exec('echo "pizza" | pachctl auth use-auth-token');
  cy.deleteReposAndPipelines();
});

describe('landing page', () => {
  after(() => {
    cy.exec('pachctl auth set cluster projectCreator allClusterUsers');
  });
  it('enables project actions when you have the correct role', () => {
    cy.multiLineExec(`
      pachctl auth set cluster projectCreator allClusterUsers
      pachctl auth set cluster none user:kilgore@kilgore.trout
      pachctl create project test
      pachctl auth set project test projectOwner user:kilgore@kilgore.trout
    `);

    // create project
    cy.login();
    cy.findByRole('button', {name: /create project/i}).should('be.enabled');

    // delete project
    cy.findByRole('button', {
      name: /test overflow menu/i,
    }).click();

    cy.findByRole('menuitem', {
      name: /delete project/i,
    }).should('exist');

    // edit project
    cy.findByRole('menuitem', {
      name: /edit project info/i,
    }).should('exist');

    // edit project roles
    cy.findByRole('menuitem', {
      name: /edit project roles/i,
    }).should('exist');
  });

  it('disables project actions when you do not have the correct role', () => {
    cy.multiLineExec(`
      pachctl auth set cluster none allClusterUsers
      pachctl auth set cluster none user:kilgore@kilgore.trout
      pachctl create project test
      pachctl auth set project test none user:kilgore@kilgore.trout
    `);

    cy.login();
    cy.viewAllLandingPageProjects();

    cy.findByRole('row', {
      name: /test/i,
    }); // wait for APIs to finish so that the button will be properly not showing

    // create project
    cy.findByRole('button', {name: /create project/i}).should('not.exist');

    // delete project
    cy.findByRole('button', {
      name: /test overflow menu/i,
    }).click();

    // Ensures menu has loaded before asserting
    cy.findByRole('button', {
      name: /test overflow menu/i,
    });

    cy.findByRole('menuitem', {
      name: /delete project/i,
    }).should('not.exist');

    // edit project
    cy.findByRole('menuitem', {
      name: /edit project info/i,
    }).should('not.exist');

    // view roles
    cy.findByRole('menuitem', {
      name: /view project roles/i,
    }).should('exist');
  });
});
