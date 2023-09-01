beforeEach(() => {
  cy.exec('echo "pizza" | pachctl auth use-auth-token');
  cy.deleteReposAndPipelines().logout();
});

after(() => {
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

describe('Repo / Pipeline table and DAG view', () => {
  it('displays no repos when I have neither projectViewer nor repoReader on any repo within the project', () => {
    cy.multiLineExec(`
      echo "pizza" | pachctl auth use-auth-token
      pachctl create project root --description "Project made by pach:root. You can see nothing."
      pachctl create repo repo1 --project root
      echo '{"pipeline":{"name":"pipeline1","project":{"name":"root"}},"description":"A sleepy pipeline.","input":{"pfs":{"glob":"/*","repo":"repo1","project":"root"}},"transform":{"cmd":["sh"],"stdin": ["sleep 0"]}}' | pachctl create pipeline
      pachctl auth set cluster none user:kilgore@kilgore.trout
      `);

    cy.login();

    cy.findByRole('button', {name: /view project root/i}).click();

    // create repo on sidebar
    cy.findByRole('link', {name: 'DAG'}).should('exist'); // Ensure item is loaded
    cy.findByRole('button', {name: /create repo/i}).should('not.exist');

    // create repo button on empty dag
    cy.findByText(/If you are more familiar with/i); // Ensure item is loaded
    cy.findByRole('button', {name: 'Create Your First Repo'}).should(
      'be.disabled',
    );

    // Pipelines table
    cy.findByRole('link', {name: /pipelines/i}).click();

    cy.findByRole('heading', {
      name: /this dag has no pipelines/i,
    });

    // Repos table
    cy.findByRole('link', {name: /repositories/i}).click();

    cy.findByRole('heading', {
      name: /this dag has no repositories/i,
    });
  });

  it('displays create repo buttons when you have projectWriter', () => {
    cy.multiLineExec(`
      echo "pizza" | pachctl auth use-auth-token
      pachctl create project testProjectWriter
      pachctl auth set cluster none user:kilgore@kilgore.trout
      pachctl auth set project testProjectWriter projectWriter user:kilgore@kilgore.trout
    `);

    cy.login('/lineage/testProjectWriter');

    // create repo on sidebar
    cy.findByRole('button', {name: /create repo/i}).should('exist');

    // create repo button on empty dag
    cy.findByRole('button', {name: 'Create Your First Repo'}).should(
      'be.enabled',
    );
  });

  it('displays one repo as reader and one as None when I have projectViewer and repoReader on one of two repos.', () => {
    cy.multiLineExec(`
      echo "pizza" | pachctl auth use-auth-token
      pachctl create project viper --description "Kilgore has projectViewer. They have repoReader on one of the two repos."
      pachctl create repo repo1 --project viper
      pachctl create repo repo2 --project viper
      pachctl create repo isolatedRepo --project viper
      echo '{"pipeline":{"name":"pipeline1","project":{"name":"viper"}},"description":"A sleepy pipeline.","input":{"pfs":{"glob":"/*","repo":"repo1","project":"viper"}},"transform":{"cmd":["sh"],"stdin": ["sleep 0"]}}' | pachctl create pipeline
      echo '{"pipeline":{"name":"pipeline2","project":{"name":"viper"}},"description":"A sleepy pipeline.","input":{"pfs":{"glob":"/*","repo":"repo2","project":"viper"}},"transform":{"cmd":["sh"],"stdin": ["sleep 0"]}}' | pachctl create pipeline
      pachctl auth set cluster none user:kilgore@kilgore.trout
      pachctl auth set project viper projectViewer user:kilgore@kilgore.trout
      pachctl auth set repo repo1 repoReader user:kilgore@kilgore.trout --project viper
      pachctl auth set repo isolatedRepo repoReader user:kilgore@kilgore.trout --project viper
      pachctl auth set repo pipeline1 repoReader user:kilgore@kilgore.trout --project viper
    `);

    // DAG view
    cy.login('/lineage/viper/repos/repo2');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/None/i);
    });

    cy.visit('/lineage/viper/repos/isolatedRepo');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoReader/i);
      cy.findByRole('button', {
        name: /delete/i,
      })
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoOwner to delete this.');

    cy.visit('/lineage/viper/repos/repo1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoReader/i);
      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.disabled');
    });

    cy.visit('/lineage/viper/repos/pipeline1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoReader/i);
      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.disabled');
    });

    cy.visit('/lineage/viper/pipelines/pipeline1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoReader/i);
    });

    // Pipelines table
    cy.findByRole('link', {name: /pipelines/i}).click();
    cy.findByRole('button', {
      name: /pipeline1/i,
    }).within(() => {
      cy.findByRole('cell', {name: /reporeader/i});
      cy.findByTestId('DropdownButton__button').click();
      cy.findByRole('menuitem', {name: /see all roles/i}).click();
    });
    cy.findByRole('heading', {
      name: 'Repo Level Roles: viper/pipeline1',
    }).should('exist');
    cy.findByRole('dialog').within(() => {
      cy.findByRole('button', {name: /done/i}).click();
    });

    cy.findByRole('button', {
      name: /pipeline2/i,
    })
      .should(($row) => {
        expect($row.attr('aria-disabled')).to.equal('true');
      })
      .within(() => {
        cy.findByRole('cell', {name: /None/i});
      });

    // Repos table
    cy.findByRole('link', {name: /repositories/i}).click();
    cy.findByRole('button', {
      name: /repo1/i,
    }).within(() => {
      cy.findByRole('cell', {name: /reporeader/i});
      cy.findByTestId('DropdownButton__button').click();
      cy.findByRole('menuitem', {name: /see all roles/i}).click();
    });
    cy.findByRole('heading', {
      name: 'Repo Level Roles: viper/repo1',
    }).should('exist');
    cy.findByRole('dialog').within(() => {
      cy.findByRole('button', {name: /done/i}).click();
    });

    cy.findByRole('button', {
      name: /repo2/i,
    })
      .should(($row) => {
        expect($row.attr('aria-disabled')).to.equal('true');
      })
      .within(() => {
        cy.findByRole('cell', {name: /None/i});
      });
  });

  it('disables actions when you do not have the correct role on list views', () => {
    cy.multiLineExec(`
      echo "pizza" | pachctl auth use-auth-token
      pachctl create project hornet --description "Kilgore has repoReader, repoWriter, and repoOwner across three repos."
      pachctl create repo repo1-reader --project hornet
      pachctl create repo repo2-writer --project hornet
      pachctl create repo repo3-owner --project hornet
      pachctl auth set cluster none user:kilgore@kilgore.trout
      pachctl auth set project hornet projectViewer user:kilgore@kilgore.trout
      pachctl auth set repo repo1-reader repoReader user:kilgore@kilgore.trout --project hornet
      pachctl auth set repo repo2-writer repoWriter user:kilgore@kilgore.trout --project hornet
      pachctl auth set repo repo3-owner repoOwner user:kilgore@kilgore.trout --project hornet
      pachctl create repo repo4 --project hornet
      pachctl create repo repo5 --project hornet
      pachctl create repo repo6 --project hornet
      echo '{"pipeline":{"name":"pipeline1-reader","project":{"name":"hornet"}},"description":"A sleepy pipeline.","input":{"pfs":{"glob":"/*","repo":"repo4","project":"hornet"}},"transform":{"cmd":["sh"],"stdin": ["sleep 0"]}}' | pachctl create pipeline
      echo '{"pipeline":{"name":"pipeline2-writer","project":{"name":"hornet"}},"description":"A sleepy pipeline.","input":{"pfs":{"glob":"/*","repo":"repo5","project":"hornet"}},"transform":{"cmd":["sh"],"stdin": ["sleep 0"]}}' | pachctl create pipeline
      echo '{"pipeline":{"name":"pipeline3-owner","project":{"name":"hornet"}},"description":"A sleepy pipeline.","input":{"pfs":{"glob":"/*","repo":"repo6","project":"hornet"}},"transform":{"cmd":["sh"],"stdin": ["sleep 0"]}}' | pachctl create pipeline
      pachctl auth set repo pipeline1-reader repoReader user:kilgore@kilgore.trout --project hornet
      pachctl auth set repo pipeline2-writer repoWriter user:kilgore@kilgore.trout --project hornet
      pachctl auth set repo pipeline3-owner repoOwner user:kilgore@kilgore.trout --project hornet
      `);

    // source repos
    cy.login('/lineage/hornet/repos/repo1-reader');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoReader/i);

      cy.findByRole('button', {
        name: /upload files/i,
      })
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoWriter to upload files.');

    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByRole('button', {
        name: /delete/i,
      })
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoOwner to delete this.');

    cy.visit('/lineage/hornet/repos/repo2-writer');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoWriter/i);

      cy.findByRole('link', {
        name: /upload files/i,
      });

      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.disabled');

      cy.findByRole('button', {
        name: /see all roles/i,
      }).click();
    });
    cy.findByRole('heading', {
      name: 'Repo Level Roles: hornet/repo2-writer',
    }).should('exist');
    cy.findByRole('dialog').within(() => {
      cy.findByRole('button', {name: /done/i}).click();
    });

    cy.visit('/lineage/hornet/repos/repo3-owner');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoOwner/i);

      cy.findByRole('link', {
        name: /upload files/i,
      });

      cy.findByRole('button', {
        name: /delete/i,
      });

      cy.findByRole('button', {
        name: /set roles/i,
      }).click();
    });
    cy.findByRole('heading', {
      name: 'Set Repo Level Roles: hornet/repo3-owner',
    }).should('exist');
    cy.findByRole('dialog').within(() => {
      cy.findByRole('button', {name: /done/i}).click();
    });

    // pipelines
    cy.visit('/lineage/hornet/pipelines/pipeline1-reader');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoReader/i);
      cy.findByRole('button', {
        name: /delete/i,
      })
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoOwner to delete this.');

    cy.visit('/lineage/hornet/pipelines/pipeline2-writer');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoWriter/i);
      cy.findByRole('button', {
        name: /delete/i,
      })
        .should('be.disabled')
        .trigger('mouseenter', {force: true});

      cy.findByRole('button', {
        name: /see all roles via repo/i,
      }).click();
    });
    cy.findByRole('heading', {
      name: 'Repo Level Roles: hornet/pipeline2-writer',
    }).should('exist');
    cy.findByRole('dialog').within(() => {
      cy.findByRole('button', {name: /done/i}).click();
    });
    cy.findByText('You need at least repoOwner to delete this.');

    cy.visit('/lineage/hornet/pipelines/pipeline3-owner');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoOwner/i);
      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.enabled');

      cy.findByRole('button', {
        name: /set roles via repo/i,
      }).click();
    });
    cy.findByRole('heading', {
      name: 'Set Repo Level Roles: hornet/pipeline3-owner',
    }).should('exist');
    cy.findByRole('dialog').within(() => {
      cy.findByRole('button', {name: /done/i}).click();
    });

    // output repos
    cy.visit('/lineage/hornet/repos/pipeline1-reader');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoReader/i);

      cy.findByRole('button', {
        name: /upload files/i,
      })
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoWriter to upload files.');

    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByRole('button', {
        name: /delete/i,
      })
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoOwner to delete this.');

    cy.visit('/lineage/hornet/repos/pipeline2-writer');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoWriter/i);

      cy.findByRole('link', {
        name: /upload files/i,
      });

      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.disabled');
    });

    cy.visit('/lineage/hornet/repos/pipeline3-owner');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoOwner/i);

      cy.findByRole('link', {
        name: /upload files/i,
      });

      cy.findByRole('button', {
        name: /delete/i,
      });
    });
  });

  it('displays clusterAdmin and all repo roles when you have them', () => {
    cy.multiLineExec(`
      echo "pizza" | pachctl auth use-auth-token
      pachctl create project raptor --description "Kilgore has projectViewer. Kilgore has repoReader on one of the two repos."
      pachctl create repo repo1 --project raptor
      echo '{"pipeline":{"name":"pipeline1","project":{"name":"raptor"}},"description":"A sleepy pipeline.","input":{"pfs":{"glob":"/*","repo":"repo1","project":"raptor"}},"transform":{"cmd":["sh"],"stdin": ["sleep 0"]}}' | pachctl create pipeline
      pachctl create repo isolatedRepo --project raptor
      pachctl auth set cluster clusterAdmin user:kilgore@kilgore.trout
      pachctl auth set project raptor projectOwner,projectWriter,projectViewer user:kilgore@kilgore.trout
      pachctl auth set repo repo1 repoOwner,repoWriter,repoReader user:kilgore@kilgore.trout --project raptor
      pachctl auth set repo isolatedRepo repoOwner,repoWriter,repoReader user:kilgore@kilgore.trout --project raptor
      pachctl auth set repo pipeline1 repoOwner,repoWriter,repoReader user:kilgore@kilgore.trout --project raptor
    `);

    // DAG view
    cy.login('/lineage/raptor/repos/repo1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(
        /clusterAdmin, projectOwner, repoOwner, repoReader, repoWriter/i,
      );
      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.disabled');
    });

    cy.visit('/lineage/raptor/repos/isolatedRepo');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(
        /clusterAdmin, projectOwner, repoOwner, repoReader, repoWriter/i,
      );
      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.enabled');
    });

    cy.visit('/lineage/raptor/repos/pipeline1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(
        /clusterAdmin, projectOwner, repoOwner, repoReader, repoWriter/i,
      );
      cy.findByRole('button', {
        name: /delete/i,
      }).should('be.disabled');
    });

    cy.visit('/lineage/raptor/pipelines/pipeline1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(
        /clusterAdmin, projectOwner, repoOwner, repoReader, repoWriter/i,
      );
    });

    // Pipelines table
    cy.findByRole('link', {name: /pipelines/i}).click();

    cy.findByRole('button', {
      name: /pipeline1/i,
    }).within(() => {
      cy.findByRole('cell', {
        name: /clusterAdmin, projectOwner, repoOwner, repoReader, repoWriter/i,
      });
    });

    // Repos table
    cy.findByRole('link', {name: /repositories/i}).click();

    cy.findByRole('button', {
      name: /repo1/i,
    }).within(() => {
      cy.findByRole('cell', {
        name: /clusterAdmin, projectOwner, repoOwner, repoReader, repoWriter/i,
      });
    });
  });
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
