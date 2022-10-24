describe('Community Edition', () => {
    beforeEach(() => {
      cy.visit('/');
      cy.findByText('Skip tutorial', {timeout: 12000}).click();
    });

    afterEach(() => {
      cy.exec("echo y | pachctl auth deactivate", {failOnNonZeroExit: false})
        .exec("pachctl enterprise deactivate")
        .exec("pachctl license delete-all");
    });
  
    after(() => {
      cy.deleteReposAndPipelines();
    });
  
    it('should show a community edition banner', () => {
      cy.findByText('Community Edition').should('exist');
      cy.findByText('Pipelines: 0/16').should('exist');

      cy
      .exec("jq -r .pachReleaseCommit version.json")
      .then((res) => {
        cy.exec("pachctl create repo images")
          .exec(
            `pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/${res.stdout}/examples/opencv/edges.json`
          )})

      cy.findByText('Pipelines: 1/16', {timeout: 12000}).should('exist');
    });
  
    it('should remove the community edition banner when an enterprise license is entered', () => {
      cy.findByText('Community Edition').should('exist');
      cy.findByText('Upgrade to Enterprise').parent().parent().should('have.attr', 'href', 'https://www.pachyderm.com/trial-console/?utm_source=console')

  
      cy.exec("echo $PACHYDERM_ENTERPRISE_KEY | pachctl license activate");
      
      cy.reload();
      cy.findByText('Project Preview', {timeout: 12000}).should('exist');
      cy.findByText('Community Edition').should('not.exist');
    });
  });
