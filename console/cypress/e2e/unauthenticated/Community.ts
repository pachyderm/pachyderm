describe('Community Edition Banner', () => {
  describe('In Community Edition', () => {
    beforeEach(() => {
      cy.deleteReposAndPipelines();
      cy.visit('/');
    });

    after(() => {
      cy.deleteReposAndPipelines();
    });

    it('should show a community edition banner', () => {
      cy.findByText('Community Edition').should('exist');
      cy.findByRole('link', {
        name: 'Upgrade to Enterprise',
      }).should(
        'have.attr',
        'href',
        'https://www.pachyderm.com/trial-console/?utm_source=console',
      );

      cy.findByText('Pipelines: 0/16').should('exist');

      cy.exec('jq -r .pachReleaseCommit version.json').then((res) => {
        cy.exec('pachctl create repo images').exec(
          `pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/${res.stdout}/examples/opencv/edges.pipeline.json`,
        );
      });

      cy.findByText('Pipelines: 1/16', {timeout: 12000}).should('exist');
    });
  });
});
