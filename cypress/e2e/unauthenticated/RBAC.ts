before(() => {
  cy.deleteReposAndPipelines();
});

after(() => {
  cy.deleteReposAndPipelines();
});

describe('landing page', () => {
  it('project actions are enabled', () => {
    // create project
    cy.visit('/');
    cy.findByRole('button', {name: /create project/i}).should('be.enabled');

    // delete project
    cy.findByRole('button', {
      name: /default overflow menu/i,
    }).click();

    cy.findByRole('menuitem', {
      name: /delete project/i,
    }).should('exist');

    // edit project
    cy.findByRole('menuitem', {
      name: /edit project info/i,
    }).should('exist');

    // edit / view project roles
    cy.findByRole('menuitem', {
      name: /edit project roles/i,
    }).should('not.exist');
    cy.findByRole('menuitem', {
      name: /view project roles/i,
    }).should('not.exist');
  });
});

describe('project view', () => {
  it('project rbac actions are disabled', () => {
    cy.visit('/');
    cy.findByRole('button', {name: /view project default/i}).click();

    cy.findByRole('button', {name: /create repo/i}).should('exist');
    cy.findByRole('button', {name: /user roles/i}).should('not.exist');
    cy.findByRole('button', {name: 'Create Your First Repo'}).should(
      'be.enabled',
    );
  });
});

describe('Repo / Pipeline table', () => {
  it('roles column does not appear in the repos nor pipelines table when auth is not enabled', () => {
    cy.multiLineExec(`
      pachctl create project viper
      pachctl create repo repo1 --project viper
      pachctl create repo isolatedRepo --project viper
      echo '{"pipeline":{"name":"pipeline1","project":{"name":"viper"}},"description":"A sleepy pipeline.","input":{"pfs":{"glob":"/*","repo":"repo1","project":"viper"}},"transform":{"cmd":["sh"],"stdin": ["sleep 0"]}}' | pachctl create pipeline
    `);

    // DAG view
    cy.visit('/lineage/viper/repos/repo1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByRole('status', {timeout: 15000}).should('not.exist'); // Wait for details to load
      cy.findByText(/your roles/).should('not.exist');
      cy.findByRole('link', {
        name: /upload files/i,
      });
      cy.findByRole('button', {
        name: /delete/i,
      })
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText(
      "This repo can't be deleted while it has associated pipelines.",
    );

    cy.visit('/lineage/viper/repos/pipeline1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByRole('status', {timeout: 15000}).should('not.exist'); // Wait for details to load
      cy.findByText(/your roles/).should('not.exist');
      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.disabled');
    });

    cy.visit('/lineage/viper/repos/isolatedRepo');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByRole('status', {timeout: 15000}).should('not.exist'); // Wait for details to load
      cy.findByText(/your roles/).should('not.exist');
      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.enabled');
    });

    cy.visit('/lineage/viper/pipelines/pipeline1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByRole('status', {timeout: 15000}).should('not.exist'); // Wait for details to load
      cy.findByText(/your roles/).should('not.exist');
      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.enabled');
    });

    // Pipelines table
    cy.findByRole('link', {name: /pipelines/i}).click();
    cy.findByRole('status', {timeout: 15000}).should('not.exist'); // Wait for details to load

    cy.findByRole('cell', {
      name: /pipeline1/i,
    });

    cy.findByRole('columnheader', {
      name: /roles/i,
    }).should('not.exist');

    cy.findByRole('button', {
      name: /pipeline1/i,
    }).within(() => {
      cy.findByRole('cell', {name: /none/i}).should('not.exist');
      cy.findByTestId('DropdownButton__button').click();
      cy.findByRole('menuitem', {name: /roles/i}).should('not.exist');
    });

    // Repos table
    cy.findByRole('link', {name: /repositories/i}).click();

    cy.findByRole('cell', {
      name: /repo1/i,
    });

    cy.findByRole('columnheader', {
      name: /roles/i,
    }).should('not.exist');

    cy.findByRole('button', {
      name: /repo1/i,
    }).within(() => {
      cy.findByRole('cell', {name: /none/i}).should('not.exist');
      cy.findByTestId('DropdownButton__button').click();
      cy.findByRole('menuitem', {name: /roles/i}).should('not.exist');
    });
  });
});
