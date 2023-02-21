describe('Onboarding', () => {
  before(() => {
    cy.exec(`pachctl delete repos --all -f --project onboarding`, {
      failOnNonZeroExit: false,
    });
    cy.exec('pachctl create project onboarding', {
      failOnNonZeroExit: false,
    });
  });

  after(() => {
    cy.exec(`pachctl delete pipelines --all -f --project onboarding`, {
      failOnNonZeroExit: false,
    });
    cy.exec(`pachctl delete repos --all -f --project onboarding`, {
      failOnNonZeroExit: false,
    });
    cy.exec('pachctl delete project onboarding -f', {
      failOnNonZeroExit: false,
    });
  });

  beforeEach(() => {
    cy.visit('/');
    cy.findByText('Projects');
  });

  it('when a DAG is empty, I can create and view a repo.', () => {
    cy.visit('/lineage/onboarding');

    // DAG view
    cy.findByRole('button', {
      name: /create your first repo/i,
    }).click();

    // Create Repo Modal
    cy.findByRole('dialog', {timeout: 12000}).within(() => {
      cy.findByRole('textbox', {
        name: /name/i,
      })
        .clear()
        .type('NewRepo');

      cy.findByRole('textbox', {
        name: /description \(optional\)/i,
      }).type('A repo');

      cy.findByRole('button', {
        name: /create/i,
      }).click();
    });

    // DAG view
    cy.findByText('NewRepo').click();

    // Side panel
    cy.findByRole('tab', {
      name: /info/i,
    }).click();
    cy.findByText('A repo', {timeout: 15000}).should('be.visible');
  });
});
