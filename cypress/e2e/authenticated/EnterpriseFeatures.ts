// Any tests that need an enterprise license to work.
describe('Enterprise Features', () => {
  describe('Dag Node', () => {
    // NOTE: This test must live here because we must have an enterprise key to
    // be able to set parallelism to a number more than 8.

    before(() => {
      cy.exec('echo "pizza" | pachctl auth use-auth-token');
      cy.deleteReposAndPipelines();
      cy.login();
      cy.exec(`
    pachctl create repo data
    echo '{"pipeline": {"name": "workers"}, "parallelismSpec": {"constant": 999} ,"transform": {"cmd": ["sh"],"stdin": [""]},"input": {"pfs": {"glob": "/*","repo": "data"}}}'  | pachctl create pipeline
    `);
    });

    beforeEach(() => {
      cy.visit('/lineage/default');
    });
    after(() => {
      cy.exec('echo "pizza" | pachctl auth use-auth-token');
      cy.deleteReposAndPipelines();
    });

    it('should show truncate parallelism info in footer of dag node', () => {
      cy.get(
        'g[id="GROUP_e2dc1146a182e5c1d924aab7badb60b37f7538c1_parallelism"]',
      )
        .should('exist')
        .findByText('>100');
    });
  });
});
