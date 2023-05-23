describe('Landing', () => {
  before(() => {
    cy.deleteReposAndPipelines();
    cy.setupProject();
    cy.exec('pachctl delete project new-project', {failOnNonZeroExit: false});
  });
  beforeEach(() => {
    cy.visit('/');
  });

  after(() => {
    cy.deleteReposAndPipelines();
    cy.exec('pachctl delete project new-project', {failOnNonZeroExit: false});
  });

  it('should show default project info', () => {
    cy.contains('[role="row"]', /default/i).click();
    cy.findByRole('heading', {
      name: 'Project Preview',
    });

    cy.findByText('Total No. of Repos/Pipelines');
    cy.findByText('2/1');

    cy.findByText('Total Data Size');

    cy.findByText('Pipeline Status');
    cy.findByText('Last Job');
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
      })
        .should('have.value', 'New desc')
        .clear()
        .type('Edit desc');

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
});
