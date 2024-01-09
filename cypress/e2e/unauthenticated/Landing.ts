describe('Landing', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.setupProject();
    cy.exec('pachctl delete project new-project', {failOnNonZeroExit: false});
    cy.exec(
      `echo '{"createPipelineRequest":{"resourceRequests":{"cpu":1,"memory":"256Mi","disk":"1Gi"},"sidecarResourceRequests":{"cpu":1,"memory":"256Mi","disk":"1Gi"}}}' | pachctl update defaults --cluster`,
    );
  });
  beforeEach(() => {
    cy.visit('/');
  });

  after(() => {
    cy.deleteReposAndPipelines();
    cy.exec('pachctl delete project new-project', {failOnNonZeroExit: false});
    cy.exec(
      `echo '{"createPipelineRequest":{"resourceRequests":{"cpu":1,"memory":"256Mi","disk":"1Gi"},"sidecarResourceRequests":{"cpu":1,"memory":"256Mi","disk":"1Gi"}}}' | pachctl update defaults --cluster`,
    );
  });

  it('should create a new project, edit its description, then delete it', () => {
    // create new project
    cy.findByRole('button', {
      name: /create project/i,
      timeout: 12000,
    }).click();

    cy.findByRole('dialog', {timeout: 12000}).within(() => {
      cy.findByRole('textbox', {
        name: /name/i,
        exact: false,
      }).type('new-project');

      cy.findByRole('textbox', {
        name: /description/i,
        exact: false,
      }).type('New desc');

      cy.findByRole('button', {
        name: /create/i,
      }).click();
    });

    cy.findByRole('heading', {
      name: /new-project/i,
      timeout: 15000,
    });

    cy.contains('[role="row"]', /new-project/i).within(() => {
      // because we run cypress at a small screen size, the text will be hidden
      cy.findByText('New desc').should('not.be.visible');
    });

    // edit project description
    cy.findByRole('button', {
      name: /new-project overflow menu/i,
    }).click();

    cy.findByRole('menuitem', {
      name: /edit project info/i,
    }).click();

    cy.findByRole('dialog', {timeout: 12000}).within(() => {
      cy.findByRole('textbox', {
        name: /description/i,
        exact: false,
      }).should('have.value', 'New desc');
      cy.findByRole('textbox', {
        name: /description/i,
        exact: false,
      }).clear();
      cy.findByRole('textbox', {
        name: /description/i,
        exact: false,
      }).type('Edit desc');

      cy.findByRole('button', {
        name: /confirm changes/i,
      }).click();
    });

    cy.findByRole('dialog').should('not.exist');

    cy.contains('[role="row"]', /new-project/i).within(() => {
      // because we run cypress at a small screen size, the text will be hidden
      cy.findByText('Edit desc').should('not.be.visible');
    });

    // delete project
    cy.findByRole('button', {
      name: /new-project overflow menu/i,
    }).click();

    cy.findByRole('menuitem', {
      name: /delete project/i,
    }).click();

    cy.findByRole('dialog', {timeout: 12000}).within(() => {
      cy.findByRole('textbox').should('have.value', '').type('new-project');

      cy.findByRole('button', {
        name: /delete project/i,
      }).click();
    });

    cy.findByRole('dialog').should('not.exist');

    cy.contains('[role="row"]', /new-project/i).should('not.exist');
  });

  it('should update and apply the cluster config and regenerate a pipeline', () => {
    cy.findByRole('button', {
      name: /cluster defaults/i,
      timeout: 12000,
    }).click();

    cy.findByRole('textbox').should('contain.text', 'createPipelineRequest');
    cy.findByRole('textbox').clear();
    cy.findByRole('textbox').type(`{
      "createPipelineRequest": {
          "resourceRequests": {
              "cpu": 1,
              "memory": "128Mi",
              "disk": "1Gi"`);

    cy.findByRole('button', {
      name: /continue/i,
    }).click();

    cy.findByRole('radio', {
      name: /save cluster defaults and regenerate pipelines/i,
    }).click();

    cy.findByText('1 pipeline will be affected');

    cy.findByRole('button', {
      name: /save/i,
    }).click();

    cy.findByRole('alert').should('have.text', 'Cluster defaults saved successfuly');

    cy.visit('/lineage/default/pipelines/edges/spec');

    // new setting has been applied
    cy.findByText(`"128Mi"`, {
      timeout: 12000,
    });
  });

  it('should update and apply a project config and regenerate a pipeline', () => {
    cy.findByRole('button', {
      name: /default overflow menu/i,
    }).click();

    cy.findByRole('menuitem', {
      name: /edit project defaults/i,
    }).click();

    cy.findByRole('textbox').should('contain.text', 'createPipelineRequest');
    cy.findByRole('textbox').clear();
    cy.findByRole('textbox').type(`{
      "createPipelineRequest": {
          "resourceRequests": {
              "cpu": 1,
              "memory": "128Mi",
              "disk": "2Gi"`);

    cy.findByRole('button', {
      name: /continue/i,
    }).click();

    cy.findByRole('radio', {
      name: /save project defaults and regenerate pipelines/i,
    }).click();

    cy.findByText('1 pipeline will be affected');

    cy.findByRole('button', {
      name: /save/i,
    }).click();

    cy.findByRole('alert').should('have.text', 'Project defaults saved successfuly');

    cy.visit('/lineage/default/pipelines/edges/spec');

    // new setting has been applied
    cy.findByText(`"2Gi"`, {
      timeout: 12000,
    });
  });
});
