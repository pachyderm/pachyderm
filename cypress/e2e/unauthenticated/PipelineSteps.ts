describe('Pipelines', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.setupProject('error-opencv').visit('/');
    cy.exec('pachctl create repo images2');
  });

  beforeEach(() => {
    cy.visit('/');
    cy.findAllByText(/^View(\sProject)*$/)
      .eq(0)
      .click();
  });

  after(() => {
    // Delete repos and pipelines can trigger an error as the browser and polling will remain open
    // while they are deleted. Therefore we can visit about:blank to stop an error from happening.
    cy.window().then((win) => {
      win.location.href = 'about:blank';
    });
    cy.deleteReposAndPipelines();
  });

  it('should allow a user to select a subset of pipelines to inspect jobs and apply a global ID', () => {
    cy.get('#GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65').should('exist');
    cy.get('#GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838').should('exist');
    cy.get('#GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6').should('exist');

    cy.findByText('Pipelines').click();
    cy.findAllByTestId('PipelineListRow__row').should('have.length', 2);

    cy.findAllByTestId('PipelineListRow__row').eq(1).click();
    cy.findByText('Detailed info for 1 pipeline');
    cy.findByRole('tab', {name: 'Jobs'}).click();

    cy.findAllByTestId('JobsList__row').should('have.length', 1);

    cy.findAllByTestId('JobsList__row')
      .first()
      .within(() => cy.findByTestId('DropdownButton__button').click());

    cy.findAllByText('Apply Global ID and view in DAG').click();

    cy.get('#GROUP_8deb3fe2e77d2fe21f5825ac5e34951ac4eb8e65').should('exist');
    cy.get('#GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838').should('exist');
    cy.get('#GROUP_52faf83dd0fff4b0d510f5326e2bf66e8b5a2ed6').should(
      'not.exist',
    );

    // table filters are kept on a back nav
    cy.go('back');
    cy.findAllByTestId('JobsList__row').should('have.length', 1);
  });

  it('should allow a user to create a pipeline', () => {
    cy.findByRole('button', {name: 'Create', timeout: 12000}).click();
    cy.findByRole('menuitem', {
      name: 'Pipeline',
      timeout: 12000,
    }).click();

    cy.findByRole('status').should('not.exist');

    cy.findByRole('textbox').should('contain.text', 'pipeline');
    cy.findByRole('textbox').type(
      `{selectAll}{backspace}{
  "pipeline": {
    "name": "edges2"
  },
  "description": "A pipeline that performs images edge detection by using the OpenCV library.",
  "input": {
    "pfs": {
      "glob": "/*",
      "repo": "images2"
    }
  },
  "transform": {
    "cmd": ["python3", "/edges.py"],
    "image": "pachyderm/opencv:1.0"
  }
}${'{del}'.repeat(40)}`,
      {delay: 0},
    );
    cy.findByRole('alert').should('not.exist');
    cy.findByRole('button', {
      name: /prettify/i,
    }).click();
    cy.findByRole('button', {
      name: /create pipeline/i,
    }).should('be.enabled');

    // assert that the effective spec combines the editor text and the cluster defaults
    cy.findByRole('tab', {name: /effective spec/i}).click();
    cy.findByRole('textbox').should('contain.text', 'edges2');

    cy.findByRole('button', {
      name: /create pipeline/i,
    }).click();

    // new pipeline has been created
    cy.get('#GROUP_fcd0346ddb31ee5a39ed7eb2271c659bf855737d').should('exist');
  });

  it('should infer the project name if not provided', () => {
    cy.exec('pachctl create project pipelineInferProject');
    cy.visit('/');
    cy.findAllByText(/^View(\sProject)*$/)
      .eq(1)
      .click();
    cy.findByRole('button', {name: 'Create', timeout: 12000}).click();
    cy.findByRole('menuitem', {
      name: 'Pipeline',
      timeout: 12000,
    }).click();

    cy.findByRole('textbox').type(
      `{selectAll}{backspace}{
  "pipeline": {
    "name": "noProject"
  },
  "transform": {}
}${'{del}'.repeat(40)}`,
      {delay: 0},
    );

    // assert that the effective spec added the correct project name
    cy.findByRole('tab', {name: /effective spec/i}).click();
    cy.findByRole('textbox').should('contain.text', 'pipelineInferProject');

    cy.findByRole('button', {
      name: /create pipeline/i,
    }).click();

    // new pipeline has been created in the same project
    cy.get('#GROUP_6e78a0c2510b56582f8b7a23aebf9e6df267512b').should('exist');
  });

  it('should allow a user to update a pipeline', () => {
    cy.findByRole('button', {
      name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
      timeout: 10000,
    }).click();
    cy.findByRole('button', {
      name: /pipeline actions/i,
    }).click();
    cy.findByRole('menuitem', {
      name: /update pipeline/i,
    }).click();

    cy.findByRole('status').should('not.exist');

    cy.findByRole('textbox').should('contain.text', 'edges');
    // update input from images to images2
    cy.findByRole('textbox').type(
      `{selectAll}{backspace}{
  "pipeline": {
    "name": "edges"
  },
  "input": {
    "pfs": {
      "glob": "/*",
      "repo": "images2"
    }
  },
  "transform": {
    "cmd": ["python3", "/edges.py"],
    "image": "pachyderm/opencv:1.0"
  }
}${'{del}'.repeat(40)}`,
      {delay: 0},
    );
    cy.findByRole('alert').should('not.exist');
    cy.findByRole('button', {
      name: /prettify/i,
    }).click();
    cy.findByRole('button', {
      name: /update pipeline/i,
    }).click();

    cy.findByRole('dialog').within(() => {
      cy.findByRole('checkbox', {
        name: /reprocess datums/i,
      }).click();
      cy.findByRole('button', {
        name: /update pipeline/i,
      }).click();
    });

    // images2 now has pipeline edges as output
    cy.findByRole('button', {
      name: 'GROUP_fa25b098cdc91b225479ff6f51bedbe124e29d5a repo',
      timeout: 10000,
    }).click();
    cy.findAllByTestId('ResourceLink__pipeline')
      .eq(0)
      .should('have.attr', 'href', '/lineage/default/pipelines/edges');
  });

  it('should allow a user to duplicate a pipeline', () => {
    cy.findByRole('button', {
      name: 'GROUP_d0e1e9a51269508c3f11c0e64c721c3ea6204838 pipeline',
      timeout: 10000,
    }).click();
    cy.findByRole('button', {
      name: /pipeline actions/i,
    }).click();
    cy.findByRole('menuitem', {
      name: /duplicate pipeline/i,
    }).click();

    cy.findByRole('status').should('not.exist');

    cy.findByRole('textbox').should('contain.text', 'edges');
    cy.findByRole('button', {
      name: /create pipeline/i,
    }).click();

    // images2 now has pipeline edges_copy as output
    cy.findByRole('button', {
      name: 'GROUP_fa25b098cdc91b225479ff6f51bedbe124e29d5a repo',
      timeout: 10000,
    }).click();
    cy.findAllByTestId('ResourceLink__pipeline')
      .eq(0)
      .should('have.attr', 'href', '/lineage/default/pipelines/edges_copy');
  });
});
