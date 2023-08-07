describe('DatumViewer', () => {
  beforeEach(() => {
    cy.deleteReposAndPipelines();
    cy.visit('/');
  });

  describe('Logs', () => {
    beforeEach(() => {
      cy.multiLineExec(`
        pachctl create repo data
        echo "gibberish" | pachctl put file data@master:badImage.png
        echo '{"pipeline": {"name": "lots-of-logs"},"description": "generate lots of logs","transform": {"cmd": ["bash"],"stdin": ["for (( i=0; i<=1500; i++)) do echo \\"log-$i\\"; done"]},"input": {"pfs": {"glob": "/*","repo": "data"}}}'  | pachctl create pipeline
      `);
    });

    after(() => {
      cy.deleteReposAndPipelines();
    });

    it('should page logs', () => {
      cy.visit('/lineage/default/pipelines/lots-of-logs');

      cy.findByRole('link', {
        name: /inspect jobs/i,
        timeout: 30000,
      }).click();

      cy.findAllByTestId('LogRow__base', {
        timeout: 30000,
      }).should('have.length.at.least', 19);

      cy.findByTestId('Pager__forward').as('forward');
      cy.findByTestId('Pager__backward').as('backward');
      cy.findByRole('button', {
        name: /refresh/i,
      }).as('refresh');

      cy.get('@forward').should('be.enabled');
      cy.get('@backward').should('be.disabled');
      cy.get('@refresh').should('be.disabled');

      cy.findAllByTestId('LogRow__base').first().parent().parent().as('page1');

      cy.findByText(/started process datum set task/).should('be.visible');

      cy.get('@forward').click();

      cy.findByText(/log-998/).should('be.visible');

      cy.get('@forward').should('be.disabled');
      cy.get('@backward').should('be.enabled');
      cy.get('@refresh').should('be.enabled');
    });
  });
});
