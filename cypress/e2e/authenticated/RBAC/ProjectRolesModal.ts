beforeEach(() => {
  cy.exec('echo "pizza" | pachctl auth use-auth-token');
  cy.deleteReposAndPipelines().logout();
});

after(() => {
  cy.exec('echo "pizza" | pachctl auth use-auth-token');
  cy.deleteReposAndPipelines();
});

describe('Project roles modal', () => {
  after(() => {
    cy.exec('pachctl auth set cluster none user:inherited@user.com');
  });

  it('should allow users to add roles from the input form', () => {
    cy.multiLineExec(`
        echo "pizza" | pachctl auth use-auth-token
        pachctl auth get-robot-token tomcat --quiet | pachctl auth use-auth-token
        pachctl create project tomcat --description "allClusterUsers have projectOwner on the project. There is one repo."
        pachctl create repo images --project tomcat
        pachctl auth set project tomcat projectOwner allClusterUsers
      `);

    cy.login();
    cy.findByRole('button', {
      name: /tomcat overflow menu/i,
    }).click();
    cy.findByRole('menuitem', {
      name: /edit project roles/i,
    }).click();

    // edit project roles modal form
    cy.findByRole('button', {
      name: 'user',
    }).click();
    cy.findByRole('menuitem', {
      name: 'robot',
    }).click();

    cy.findByRole('button', {
      name: 'projectViewer',
    }).click();

    cy.findByRole('button', {
      name: 'Add',
    }).click();

    // confirm validation error for empty input
    cy.findByRole('alert', {name: 'A name or email is required'}).should(
      'exist',
    );

    // enter a username and submit
    cy.findByRole('textbox').type('bender');

    cy.findByRole('button', {
      name: 'Add',
    }).click();

    // confirm that the rolebinding got created and the input was cleared
    cy.findByRole('alert', {name: 'A name or email is required'}).should(
      'not.exist',
    );
    cy.findByRole('textbox').should('have.value', '');
    cy.findByRole('cell', {
      name: /robot:bender/i,
    })
      .parent()
      .within(() => {
        cy.contains('projectViewer').should('exist');
      });

    // enter a different role for the same username and submit
    cy.findByRole('button', {
      name: 'projectViewer',
    }).click();
    cy.findByRole('menuitem', {
      name: 'repoOwner',
    }).click();

    cy.findByRole('textbox').type('bender');

    cy.findByRole('button', {
      name: 'Add',
    }).click();

    // confirm that adding a second role to a user keeps both roles
    cy.findByRole('cell', {
      name: /robot:bender/i,
    })
      .parent()
      .within(() => {
        cy.contains('projectOwner').should('exist');
        cy.contains('repoOwner').should('exist');
      });

    // enter a group type user role binding and submit
    cy.findByRole('button', {
      name: 'robot',
    }).click();
    cy.findByRole('menuitem', {
      name: 'group',
    }).click();

    cy.findByRole('button', {
      name: 'repoOwner',
    }).click();
    cy.findByRole('menuitem', {
      name: 'repoWriter',
    }).click();

    cy.findByRole('textbox').type('engineering');

    cy.findByRole('button', {
      name: 'Add',
    }).click();

    cy.findByRole('cell', {
      name: /group:engineering/i,
    })
      .parent()
      .within(() => {
        cy.contains('repoWriter').should('exist');
      });

    // enter a user type user role binding and submit
    cy.findByRole('button', {
      name: 'group',
    }).click();
    cy.findByRole('menuitem', {
      name: 'user',
    }).click();

    cy.findByRole('button', {
      name: 'repoWriter',
    }).click();
    cy.findByRole('menuitem', {
      name: 'projectViewer',
    }).click();

    cy.findByRole('textbox').type('user@email.com');

    cy.findByRole('button', {
      name: 'Add',
    }).click();

    cy.findByRole('cell', {
      name: /user:user@email.com/i,
    })
      .parent()
      .within(() => {
        cy.contains('projectViewer').should('exist');
      });
  });

  it('should allow users to add and remove roles from the roles table', () => {
    cy.multiLineExec(`
        echo "pizza" | pachctl auth use-auth-token
        pachctl create project tomcat
        pachctl auth set cluster none user:admin@user.com
        pachctl auth set project tomcat projectCreator user:admin@user.com
        pachctl auth set project tomcat projectViewer user:admin@user.com
        pachctl auth set cluster projectOwner user:inherited@user.com
        pachctl auth set project tomcat projectWriter user:inherited@user.com
        pachctl auth set project tomcat projectOwner user:kilgore@kilgore.trout
      `);

    cy.login();
    cy.findByRole('button', {
      name: /tomcat overflow menu/i,
    }).click();
    cy.findByRole('menuitem', {
      name: /edit project roles/i,
    }).click();

    // filter down to a specific user
    cy.findByRole('dialog').within(() => {
      cy.findAllByRole('row').should('have.length', 4);
      cy.findByRole('button', {name: /open users search/i}).click();
      cy.findByRole('searchbox').type('inherit');
      cy.findAllByRole('row').should('have.length', 2);
    });

    // add a role binding from available roles
    cy.findByRole('cell', {
      name: /user:inherited@user.com/i,
    })
      .parent()
      .within(() => {
        cy.findAllByText('repoOwner')
          .should('have.length', 1)
          .should('not.be.visible');
        cy.findAllByText('projectOwner')
          .should('have.length', 1)
          .should('be.visible');
        cy.findAllByText('projectWriter')
          .should('have.length', 1)
          .should('be.visible');
        cy.findByLabelText('add role to user:inherited@user.com').click();
        cy.findAllByText('repoOwner').click();
        cy.findAllByText('repoOwner')
          .should('have.length', 1)
          .should('be.visible');
        cy.findAllByText('projectOwner')
          .should('have.length', 1)
          .should('be.visible');
        cy.findAllByText('projectWriter')
          .should('have.length', 1)
          .should('be.visible');
      });

    // remove the role binding
    cy.findByRole('cell', {
      name: /user:inherited@user.com/i,
    })
      .parent()
      .within(() => {
        cy.contains('repoOwner').should('be.visible');
        cy.findByLabelText('delete role repoOwner').click();
        cy.contains('repoOwner').should('not.be.visible');
      });
  });

  it('should allow users to add the allClusterUsers principal', () => {
    cy.multiLineExec(`
      echo "pizza" | pachctl auth use-auth-token
      pachctl create project cluster
      pachctl auth set cluster none allClusterUsers
      pachctl auth set cluster clusterAdmin user:kilgore@kilgore.trout
    `);

    cy.login();
    cy.findByRole('button', {
      name: /cluster overflow menu/i,
    }).click();
    cy.findByRole('menuitem', {
      name: /edit project roles/i,
    }).click();

    // add allClusterUsers from new role binding form
    cy.findByRole('cell', {
      name: /allClusterUsers/i,
    }).should('not.exist');

    cy.findByRole('button', {
      name: 'user',
    }).click();
    cy.findByRole('menuitem', {
      name: 'allClusterUsers',
    }).click();

    cy.findByRole('button', {
      name: 'Add',
    }).click();

    cy.findByRole('cell', {
      name: /allClusterUsers/i,
    })
      .parent()
      .within(() => {
        cy.contains('projectViewer').should('exist');
      });

    // confirm that allClusterUsers is no longer available from the dropdown
    cy.findByRole('button', {
      name: 'user',
    }).click();
    cy.findByRole('menuitem', {
      name: 'allClusterUsers',
    }).should('not.exist');
  });

  it('should allow users to remove all roles and undo removing all roles', () => {
    cy.multiLineExec(`
      echo "pizza" | pachctl auth use-auth-token
      pachctl create project tomcat
      pachctl auth set project tomcat projectOwner user:kilgore@kilgore.trout
      pachctl auth set cluster none user:admin@user.com
      pachctl auth set project tomcat projectViewer,repoOwner user:admin@user.com
    `);

    cy.login();

    // enter the project roles modal from the DAG sidebar
    cy.findByRole('button', {name: /view project tomcat/i}).click();
    cy.findByRole('button', {name: /user roles/i}).click();

    // remove all role bindings
    cy.findByRole('cell', {
      name: /user:admin@user.com/i,
    })
      .parent()
      .within(() => {
        cy.findByText('projectViewer').should('be.visible');
        cy.findByText('repoOwner').should('be.visible');
        cy.findByRole('button', {name: 'delete all roles'}).click();
        cy.findByText('projectViewer').should('not.be.visible');
        cy.findByText('repoOwner').should('not.be.visible');
        cy.findByRole('button', {name: 'delete all roles'}).should('not.exist');
      });

    // undo removing all role bindings
    cy.findByRole('cell', {
      name: /user:admin@user.com/i,
    })
      .parent()
      .within(() => {
        cy.findByRole('button', {name: 'undo delete all roles'}).click();
      });

    cy.findByRole('button', {name: 'undo delete all roles'}).should(
      'not.exist',
    );
    cy.findByRole('cell', {
      name: /user:admin@user.com/i,
    })
      .parent()
      .within(() => {
        cy.findByRole('button', {name: 'delete all roles'}).should('exist');
        cy.findByText('projectViewer').should('exist');
        cy.findByText('repoOwner').should('exist');
        cy.findByLabelText('delete role projectViewer').should('exist');
        cy.findByLabelText('delete role repoOwner').should('exist');
      });

    // delete all roles and check that the undo option is removed if a user adds a role
    cy.findByRole('cell', {
      name: /user:admin@user.com/i,
    })
      .parent()
      .within(() => {
        cy.findByRole('button', {name: 'delete all roles'}).click();
      });
    cy.findByRole('button', {name: 'undo delete all roles'}).should('exist');
    cy.findByRole('cell', {
      name: /user:admin@user.com/i,
    })
      .parent()
      .within(() => {
        cy.findByLabelText('add role to user:admin@user.com').click();
        cy.findAllByText('repoOwner').click();
        cy.findByRole('button', {name: 'undo delete all roles'}).should(
          'not.exist',
        );
      });

    // add role from roles form and check that the undo option is removed
    cy.findByRole('cell', {
      name: /user:admin@user.com/i,
    })
      .parent()
      .within(() => {
        cy.findByRole('button', {name: 'delete all roles'}).click();
      });

    cy.findByRole('button', {name: 'undo delete all roles'}).should('exist');
    cy.findByRole('textbox').type('admin@user.com');
    cy.findByRole('button', {
      name: 'Add',
    }).click();

    cy.findByRole('cell', {
      name: /user:admin@user.com/i,
    })
      .parent()
      .within(() => {
        cy.findByRole('button', {name: 'undo delete all roles'}).should(
          'not.exist',
        );
      });
  });

  it('should show a read-only roles modal for users without projectOwner', () => {
    cy.multiLineExec(`
        echo "pizza" | pachctl auth use-auth-token
        pachctl create project read
        pachctl auth set project read projectViewer user:kilgore@kilgore.trout
        pachctl auth set cluster none allClusterUsers
        pachctl auth set cluster none user:kilgore@kilgore.trout
      `);

    cy.login();
    cy.findByRole('button', {
      name: /read overflow menu/i,
    }).click();
    cy.findByRole('menuitem', {
      name: /view project roles/i,
    }).click();

    // validate content
    cy.findByText('You need at least projectOwner to edit roles');

    cy.findByRole('cell', {
      name: /user:kilgore@kilgore.trout/i,
    })
      .parent()
      .within(() => {
        cy.contains('projectViewer').should('exist');
      });

    // validate row actions or roles form are not present
    cy.findByRole('button', {name: /delete/i}).should('not.exist');
    cy.findByRole('button', {name: /add role/i}).should('not.exist');
    cy.findByRole('textbox').should('not.exist');

    // DAG view user roles should be read only
    cy.findByRole('dialog').within(() => {
      cy.findByRole('button', {name: /done/i}).click();
    });
    cy.findByRole('button', {name: /view project read/i}).click();
    cy.findByRole('button', {name: /user roles/i}).click();
    cy.findByText('You need at least projectOwner to edit roles');
  });

  it('should instantly reflect permissions when roles are updated', () => {
    cy.multiLineExec(`
      pachctl auth set cluster none user:kilgore@kilgore.trout
      pachctl create project test
      pachctl auth set project test projectOwner user:kilgore@kilgore.trout
    `);

    cy.login();

    // permissions enabled
    cy.findByRole('button', {
      name: /test overflow menu/i,
    }).click();

    cy.findByRole('menuitem', {
      name: /delete project/i,
    }).should('exist');

    cy.findByRole('menuitem', {
      name: /edit project roles/i,
    }).click();

    // kilgore removes projectOwner from himself
    cy.findByRole('cell', {
      name: /user:kilgore@kilgore.trout/i,
    })
      .parent()
      .within(() => {
        cy.findByLabelText('delete role projectOwner').click();
      });
    cy.findByRole('dialog').within(() => {
      cy.findByRole('button', {name: /done/i}).click();
    });

    // permissions disabled
    cy.findByRole('button', {
      name: /test overflow menu/i,
    }).click();

    cy.findByRole('menuitem', {
      name: /delete project/i,
    }).should('not.exist');

    cy.findByRole('menuitem', {
      name: /view project roles/i,
    }).should('exist');
  });

  it('should filter out unwanted user types and ignored roles', () => {
    cy.multiLineExec(`
        echo "pizza" | pachctl auth use-auth-token
        pachctl create project secret
        pachctl auth set cluster projectOwner allClusterUsers
        pachctl auth set cluster none user:kilgore@kilgore.trout
        pachctl auth set cluster secretAdmin user:secretAdmin
      `);

    cy.login();
    cy.findByRole('button', {
      name: /secret overflow menu/i,
    }).click();
    cy.findByRole('menuitem', {
      name: /project roles/i,
    }).click();

    cy.findByRole('cell', {
      name: /allclusterusers/i,
    }).should('exist');

    cy.findByRole('cell', {
      name: /pach:root/i,
    }).should('not.exist');

    cy.findByRole('cell', {
      name: /internal:auth-server/i,
    }).should('not.exist');

    cy.findByRole('cell', {
      name: /user:secretAdmin/i,
    }).should('not.exist');
  });
});
