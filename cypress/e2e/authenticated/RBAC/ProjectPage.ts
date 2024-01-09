beforeEach(() => {
  cy.exec('echo "pizza" | pachctl auth use-auth-token');
  cy.deleteReposAndPipelines().logout();
});

after(() => {
  cy.exec('echo "pizza" | pachctl auth use-auth-token');
  cy.deleteReposAndPipelines();
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

  it('displays create buttons when you have projectWriter', () => {
    cy.multiLineExec(`
      echo "pizza" | pachctl auth use-auth-token
      pachctl create project testProjectWriter
      pachctl auth set cluster none user:kilgore@kilgore.trout
      pachctl auth set project testProjectWriter projectWriter user:kilgore@kilgore.trout
    `);

    cy.login('/lineage/testProjectWriter');

    // create buttons on sidebar
    cy.findByRole('button', {name: /create/i}).should('exist');

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
      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /delete repo/i})
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoOwner to delete this.');

    cy.visit('/lineage/viper/repos/repo1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoReader/i);
      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /delete repo/i}).should('be.disabled');
    });

    cy.visit('/lineage/viper/repos/pipeline1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoReader/i);
      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /delete repo/i}).should('be.disabled');
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
      cy.findByText(/reporeader/i).click({force: true});
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
      cy.findByText(/reporeader/i).click({force: true});
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

      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /upload files/i})
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoWriter to upload files.');

    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByRole('menuitem', {name: /delete repo/i})
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoOwner to delete this.');

    cy.visit('/lineage/hornet/repos/repo2-writer');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoWriter/i);

      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /upload files/i});
      cy.findByRole('menuitem', {name: /delete repo/i}).should('be.disabled');

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

      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /upload files/i});
      cy.findByRole('menuitem', {name: /delete repo/i});

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
      cy.findByRole('button', {name: 'Pipeline Actions'}).click();
      cy.findByRole('menuitem', {
        name: /delete pipeline/i,
      })
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoOwner to delete this.');

    cy.visit('/lineage/hornet/pipelines/pipeline2-writer');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoWriter/i);
      cy.findByRole('button', {name: 'Pipeline Actions'}).click();
      cy.findByRole('menuitem', {
        name: /delete pipeline/i,
      }).should('be.disabled');

      cy.findByRole('button', {
        name: /see all roles via repo/i,
      }).click({force: true});
    });
    cy.findByRole('heading', {
      name: 'Repo Level Roles: hornet/pipeline2-writer',
    }).should('exist');
    cy.findByRole('dialog').within(() => {
      cy.findByRole('button', {name: /done/i}).click();
    });

    cy.visit('/lineage/hornet/pipelines/pipeline3-owner');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoOwner/i);
      cy.findByRole('button', {name: 'Pipeline Actions'}).click();
      cy.findByRole('menuitem', {
        name: /delete pipeline/i,
      }).should('be.enabled');

      cy.findByRole('button', {
        name: /set roles via repo/i,
      }).click({force: true});
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

      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /upload files/i})
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoWriter to upload files.');

    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByRole('menuitem', {name: /delete repo/i})
        .should('be.disabled')
        .trigger('mouseenter', {force: true});
    });
    cy.findByText('You need at least repoOwner to delete this.');

    cy.visit('/lineage/hornet/repos/pipeline2-writer');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoWriter/i);

      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /upload files/i});
      cy.findByRole('menuitem', {name: /delete repo/i}).should('be.disabled');
    });

    cy.visit('/lineage/hornet/repos/pipeline3-owner');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(/repoOwner/i);

      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /upload files/i});
      cy.findByRole('menuitem', {name: /delete repo/i});
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
      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /delete repo/i}).should('be.disabled');
    });

    cy.visit('/lineage/raptor/repos/isolatedRepo');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(
        /clusterAdmin, projectOwner, repoOwner, repoReader, repoWriter/i,
      );
      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /delete repo/i}).should('be.enabled');
    });

    cy.visit('/lineage/raptor/repos/pipeline1');
    cy.findByRole('region', {name: 'project-sidebar'}).within(() => {
      cy.findByText(
        /clusterAdmin, projectOwner, repoOwner, repoReader, repoWriter/i,
      );
      cy.findByRole('button', {name: 'Repo Actions'}).click();
      cy.findByRole('menuitem', {name: /delete repo/i}).should('be.disabled');
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
